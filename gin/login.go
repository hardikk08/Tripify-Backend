package gin

import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	URL "net/url"
	"tripify-backend/config"
)

const (
	redirectUri = "http://localhost:3000/callback"
	//scope = "user-read-private user-read-email playlist-modify-private " +
	//	"user-library-read user-read-recently-played playlist-read-private playlist-modify-public" +
	//	"user-top-read user-read-playback-position user-read-recently-played"
	scope            = "user-read-email user-read-private playlist-modify-public playlist-read-private playlist-modify-private user-top-read"
	state            = "secret-key"
	grantType        = "authorization_code"
	grantTypeRefresh = "refresh_token"
)

const (
	// ErrorBinding binding error msg
	ErrorBinding string = "error binding request"
	// ErrorMerchantUUID merchantUUID error msg
	ErrorCode string = "code is required"
	// ErrorWebsiteURL websiteURL error msg
	ErrorState string = "incorrect state"
	// ErrorUnmarshalling cant unmarshall
	ErrorUnmarshalling string = "un-marshall error"
	// ErrorSpotify
	ErrorSpotify string = "error with spotify"
)

type Error struct {
	Error string `json: "error"`
}

type LoginResponse struct {
	AuthUrl string `json:"auth_url"`
}

type TokenRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type TokenResponseError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func Login(c *gin.Context) {
	spotifyUrl := fmt.Sprintf("%s/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s", config.SpotifyUrl, config.SpotifyId, redirectUri, scope, state)
	body := &LoginResponse{AuthUrl: spotifyUrl}

	c.JSON(http.StatusOK, body)
}

func GetToken(c *gin.Context) {
	url := fmt.Sprintf("%s/api/token", config.SpotifyUrl)
	var request TokenRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, Error{ErrorBinding})
		return
	}
	if request.Code == "" {
		c.JSON(http.StatusBadRequest, Error{ErrorCode})
		return
	}
	if request.State != state {
		c.JSON(http.StatusBadRequest, Error{ErrorState})
		return
	}

	params := URL.Values{}
	params.Add("code", request.Code)
	params.Add("client_id", config.SpotifyId)
	params.Add("client_secret", config.SpotifySecret)
	params.Add("redirect_uri", redirectUri)
	params.Add("grant_type", grantType)

	response, err := http.PostForm(url, params)

	defer response.Body.Close()
	reader, _ := ioutil.ReadAll(response.Body)

	switch response.StatusCode {
	case http.StatusBadRequest:
		var error TokenResponseError
		err = json.Unmarshal(reader, &error)
		c.JSON(http.StatusBadRequest, error)
	case http.StatusOK:
		var responseBody TokenResponse
		err = json.Unmarshal(reader, &responseBody)
		if err != nil {
			fmt.Println("Problem unmarshalling!")
			c.JSON(http.StatusInternalServerError, Error{ErrorUnmarshalling})
		}
		// Set cookies
		c.SetCookie("AT", responseBody.AccessToken, 3500, "/", "", false, true)
		c.SetCookie("RT", responseBody.RefreshToken, 0, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Log in successful"})
	default:
		c.JSON(http.StatusInternalServerError, Error{ErrorSpotify})
	}
}

func GetNewToken(c *gin.Context) (tokenRes TokenResponse, err error) {
	url := fmt.Sprintf("%s/api/token", config.SpotifyUrl)
	refreshToken, _ := c.Request.Cookie("RT")
	if refreshToken.Value == "" {
		return TokenResponse{}, err
	}
	params := URL.Values{}
	params.Add("client_id", config.SpotifyId)
	params.Add("client_secret", config.SpotifySecret)
	params.Add("grant_type", grantTypeRefresh)
	params.Add("refresh_token", refreshToken.Value)

	response, err := http.PostForm(url, params)

	defer response.Body.Close()
	reader, _ := ioutil.ReadAll(response.Body)

	var responseBody TokenResponse
	err = json.Unmarshal(reader, &responseBody)
	if err != nil {
		fmt.Println("Problem unmarshalling!")
		c.JSON(http.StatusInternalServerError, Error{ErrorUnmarshalling})
	}

	return responseBody, nil
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, _ := c.Request.Cookie("AT")
		if accessToken != nil && accessToken.Value != "" {
			fmt.Println(accessToken.Expires)
			fmt.Println("Access token is present, next()")
			c.Next()
		}
		refreshToken, _ := c.Request.Cookie("RT")
		if accessToken == nil && refreshToken == nil {
			fmt.Println("neither access nor refresh available, aborting...")
			c.AbortWithError(http.StatusUnauthorized, errors.New("please try to login again"))
			return
		}
		if accessToken == nil && refreshToken.Value != "" {
			fmt.Println("access token unavailable, getting new with refresh token")
			tokenRes, err := GetNewToken(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "issue with refresh token"})
			}
			// Set cookies
			c.SetCookie("AT", tokenRes.AccessToken, 3500, "/", "", false, true)
			c.AbortWithStatusJSON(http.StatusCreated, gin.H{"message": "refreshed token, send request again"})
		}
	}
}

func Logout(c *gin.Context) {
	c.SetCookie("AT", "", -1, "/", "", false, true)
	c.SetCookie("RT", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Cookies deleted!"})
}
