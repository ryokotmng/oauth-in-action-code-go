package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var resourceDetail = map[string]string{
	"name":        "Protected Resource",
	"description": "This data has been protected by OAuth 2.0",
}

//go:embed views
var clientFS embed.FS

func getAccessToken(c *gin.Context) {
	// check the auth header first
	auth := c.Request.Header.Get("authorization")
	var inToken string
	if auth != "" && strings.Contains(strings.ToLower(auth), "bearer") {
		inToken = auth
	} else if c.Request.Body != nil {
		// not in the header, check in the form body
	} else {
		inToken = ""
	}
	fmt.Printf("Incoming token: %s", inToken)
	// TODO: put the generated token into Redis
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "access_token", inToken))
}

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

func resource(c *gin.Context) {
	fmt.Println(getAccessTokenFromContext(c))
	if getAccessTokenFromContext(c) != "" {
		c.JSON(200, resourceDetail)
	} else {
		c.Error(errors.New(""))
	}
}

func getAccessTokenFromContext(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if token, ok := c.Request.Context().Value("access_token").(string); ok {
		return token
	}
	return ""
}
