package gin

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"tripify-backend/config"
)

type ProfileResponse struct {
	DisplayName string `json:"display_name"`
	Id string `json:"id"`
	Images []Image `json:"images"`
}

type Image struct {
	Height int `json:"height"`
	Width int `json:"width"`
	Url string `json:"url"`
}

func GetUserProfile(c *gin.Context) {
	url := fmt.Sprintf("%s/me", config.SpotifyApiUrl)
	// get access token
	accessToken, _ := c.Request.Cookie("AT")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken.Value))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	reader, _ := ioutil.ReadAll(res.Body)

	var response ProfileResponse
	json.Unmarshal(reader, &response)
	switch res.StatusCode {
	case http.StatusOK:
		c.JSON(http.StatusOK, response)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong!"})
	}
}
