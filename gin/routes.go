package gin

import (
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"net/http"
	"time"
)

func SetupRouter () *gin.Engine {
	router := gin.Default()
	//router.Use(cors.Default())
	router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE, PATCH",
		RequestHeaders:  "Host, Cookie, Set-Cookie, Accept, Origin, Authorization, Content-Type, Access-Control-Allow-Credentials, Set-Cookie, Cache-Control, *",
		ExposedHeaders:  "Set-Cookie, Cookie, Host, Content-Disposition, *",
		MaxAge:          1000 * time.Hour,
		Credentials:     true,
		ValidateHeaders: false, //Should be true for production. - is more secure because we validate headers as opposed to ember.
	}))
	// Healthz endpoint, used by external services for determining that
	// this API is live and currently handling requests
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "running"})
	})
	router.GET("/login", Login)
	router.POST("/token", GetToken)
	router.GET("/logout", Logout)
	// Authenticated routes
	router.GET("/user_profile", TokenAuthMiddleware(), GetUserProfile)
	router.POST("/create_playlist", TokenAuthMiddleware(), CreatePlaylist)
	router.GET("/get_top_artists", TokenAuthMiddleware(), GetTopArtistsFromSpotify)
	return router
}
