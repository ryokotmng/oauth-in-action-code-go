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
	clientId string
	clientSecret string
	redirectURIs []string
	scope string
}

var clients = []client{
	{
		clientId: "oauth-client-1",
		clientSecret: "oauth-client-secret-1",
	    redirectURIs: []string{"http://localhost:9000/callback"},
		scope: "foo bar",
	},
}

//go:embed views
var clientFS embed.FS

func main() {
	engine := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/index.html"))
	engine.SetHTMLTemplate(tmpl)

	engine.GET("/", func(c *gin.Context) {
		viewData := gin.H{
			"clients":  clients,
			"authServer":        "NONE",
		}
		c.HTML(http.StatusOK, "index.html", viewData)
	})
	engine.Run(":9001")
	fmt.Println("OAuth Authorization Server is listening at http://localhost:9000")
}
