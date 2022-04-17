package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/ryokotmng/oauth-in-action-code-go/pkg"
)

var resourceDetail = map[string]string{
	"name":        "Protected Resource",
	"description": "This data has been protected by OAuth 2.0",
}

//go:embed views
var clientFS embed.FS

func main() {
	router := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/index.html"))
	router.SetHTMLTemplate(tmpl)
	router.Use(cors.Default())
	router.Use(getAccessToken)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.POST("/resource", resource)
	router.Run(":9002")
	fmt.Println("OAuth Resource Server is listening at http://localhost:9002")
}

func getAccessToken(c *gin.Context) {
	// check the auth header first
	auth := c.Request.Header.Get("authorization")
	var inToken string
	if auth != "" && strings.Contains(strings.ToLower(auth), "bearer") {
		inToken = strings.Replace(auth, "Bearer ", "", 1)
	} else if c.Request.Body != nil {
		// not in the header, check in the form body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		}
		defer func() {
			err := c.Request.Body.Close()
			if err != nil {
				c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
			}
		}()
		type RequestBody struct {
			AccessToken string `json:"access_token"`
		}
		var b RequestBody
		json.Unmarshal(body, &b)
		inToken = b.AccessToken
	} else if c.Query("access_token") != "" {
		inToken = c.Query("access_token")
	}
	fmt.Printf("Incoming token: %s \n", inToken)

	redisClient := pkg.NewRedisClient()
	err := redisClient.Get(c, "access_token"+inToken).Err()
	if err != redis.Nil {
		fmt.Printf("We found a matching token: %s \n", inToken)
	} else {
		fmt.Println("no matching token was found.")
	}

	c.Request = pkg.SetAccessTokenKeyToContext(*c, inToken)
}

func resource(c *gin.Context) {
	fmt.Println(pkg.GetAccessTokenFromContext(c))
	if _, ok := pkg.GetAccessTokenFromContext(c); ok {
		c.JSON(200, resourceDetail)
	} else {
		c.Error(errors.New("no access token stored in context"))
	}
}
