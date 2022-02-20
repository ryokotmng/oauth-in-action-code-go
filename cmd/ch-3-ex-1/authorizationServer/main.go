package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	authorizationEndpoint = "http://localhost:9001/authorize"
	tokenEndpoint         = "http://localhost:9001/token"
)

type client struct {
	clientId     string
	clientSecret string
	redirectURIs []string
	scope        string
}

var clients = map[string]client{
	"oauth-client-1": {
		clientId:     "oauth-client-1",
		clientSecret: "oauth-client-secret-1",
		redirectURIs: []string{"http://localhost:9000/callback"},
		scope:        "foo bar",
	},
}

var codes []string

type request struct{}

var requests []request

//go:embed views
var clientFS embed.FS

func main() {
	engine := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/*.html"))
	engine.SetHTMLTemplate(tmpl)

	engine.GET("/", func(c *gin.Context) {
		viewData := gin.H{
			"clients":    clients,
			"authServer": "NONE",
		}
		c.HTML(http.StatusOK, "index.html", viewData)
	})
	engine.GET("/authorize", authorize())
	engine.POST("/approve", approve())
	engine.POST("/token", token())
	engine.Run(":9001")
	fmt.Println("OAuth Authorization Server is listening at http://localhost:9000")
}

func authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, ok := c.Params.Get("client_id")
		if !ok {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Unknown client"})
			return
		}
		cl, ok := clients[clientID]
		if !ok {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Unknown client"})
			return
		}

		viewData := gin.H{
			"client": cl, "reqid": nil, "scope": nil,
		}
		c.HTML(http.StatusOK, "approve.html", viewData)
	}
}

func approve() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func token() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
