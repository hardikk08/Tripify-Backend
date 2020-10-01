package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"tripify-backend/config"
)

type GetTopArtists struct {
	Items []Item `json:"items"`
}

type Item struct {
	ExternalUrls ExternalUrl `json:"external_urls"`
	Images       []ItemImage `json:"images"`
	Name         string      `json:"name"`
	Uri          string      `json: "uri"`
}

type ExternalUrl struct {
	Spotify string `json:"spotify"`
}

type ItemImage struct {
	Height int    `json:"height"`
	Url    string `json:"url"`
	Width  int    `json:"width"`
}

type CreatePlaylistRequest struct {
	Time  int    `json:"time"`
	Id    string `json:"id"`
	Title string `json:"title"`
}

type PlaylistResponse struct {
	Name        string      `json: "name"`
	ExternalUrl ExternalUrl `json:"external_urls"`
	Id          string      `json:"id"`
}

type TopTracksResponse struct {
	Items []Track `json:"items"`
}

type Track struct {
	Duration float64 `json:"duration_ms"`
	Uri      string `json:"uri"`
}

type URIS struct {
	Uris []string `json:"uris"`
}

func GetTopArtistsFromSpotify(c *gin.Context) {
	url := fmt.Sprintf("%s/me/top/artists", config.SpotifyApiUrl)
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

	var response GetTopArtists
	json.Unmarshal(reader, &response)
	switch res.StatusCode {
	case http.StatusOK:
		c.JSON(http.StatusOK, response)
	case http.StatusUnauthorized:
		//GetNewToken(c)
	}
}

func CreatePlaylist(c *gin.Context) {
	var totalPlaylistTimeRequired float64
	var totalPlaylistTimeAdded float64
	var playlistBody []string
	limit := 50
	var createPlaylistRequest CreatePlaylistRequest
	err := c.BindJSON(&createPlaylistRequest)
	if err != nil {
		fmt.Println("binding error")
	}
	totalPlaylistTimeRequired = float64(createPlaylistRequest.Time)
	fmt.Println(createPlaylistRequest)

	// CREATE A PLAYLIST
	playlistResponse := createPlaylist(c, createPlaylistRequest.Id, createPlaylistRequest.Title)

	// SEND TOP TRACKS TO USER & USE THEM TO CREATE A PLAYLIST
	url := fmt.Sprintf("%s/me/top/tracks?limit=%d", config.SpotifyApiUrl, limit)
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
	tracksReader, _ := ioutil.ReadAll(res.Body)
	items := TopTracksResponse{}

	json.Unmarshal(tracksReader, &items)

	// push track uris into playlist body
	for _, k := range items.Items{
		durationInSeconds := k.Duration/1000
		if totalPlaylistTimeAdded < totalPlaylistTimeRequired {
			playlistBody = append(playlistBody, k.Uri)
			totalPlaylistTimeAdded += durationInSeconds
		}
	}
	// create struct to send to spotify
	var uris URIS
	uris.Uris = playlistBody
	b, err := json.Marshal(uris)
	if err != nil {
		fmt.Println("error:", err)
	}
	urlToPushTracks := fmt.Sprintf("%s/playlists/%s/tracks", config.SpotifyApiUrl, playlistResponse.Id)
	request, err := http.NewRequest("POST", urlToPushTracks, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println(err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken.Value))
	request.Header.Set("Content-Type", "application/json")
	responseTracks, errPlaylist := http.DefaultClient.Do(request)
	if errPlaylist != nil {
		fmt.Println(errPlaylist)
	}

	defer responseTracks.Body.Close()
	playlistReader, _ := ioutil.ReadAll(responseTracks.Body)
	fmt.Println(playlistReader)
	switch responseTracks.StatusCode {
	case http.StatusOK:
		c.JSON(http.StatusOK, playlistResponse)
		totalPlaylistTimeRequired = 0
		totalPlaylistTimeAdded = 0
		playlistBody = playlistBody[:0]
	case http.StatusCreated:
		c.JSON(http.StatusOK, playlistResponse)
		totalPlaylistTimeRequired = 0
		totalPlaylistTimeAdded = 0
		playlistBody = playlistBody[:0]
	case http.StatusUnauthorized:
	}
}

func createPlaylist(c *gin.Context, userId string, title string) PlaylistResponse {
	url := fmt.Sprintf("%s/users/%s/playlists", config.SpotifyApiUrl, userId)
	accessToken, _ := c.Request.Cookie("AT")

	type CreatePlaylistRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	body := CreatePlaylistRequest{
		Name:        title,
		Description: "Created by the Tripify App",
	}

	b, err := json.Marshal(body)
	if err != nil {
		fmt.Println("error:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
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
	var response PlaylistResponse
	json.Unmarshal(reader, &response)
	return response
}
