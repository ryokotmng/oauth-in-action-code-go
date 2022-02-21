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
	router := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/index.html"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/", func(c *gin.Context) {
		viewData := gin.H{
			"clients":  clients,
			"authServer":        "NONE",
		}
		c.HTML(http.StatusOK, "index.html", viewData)
	})
	router.Run(":9001")
	fmt.Println("OAuth Authorization Server is listening at http://localhost:9000")
}
