package main

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

// authorization server information
const (
	authorizationEndpoint = "http://localhost:9001/authorize"
	tokenEndpoint         = "http://localhost:9001/token"
)

// client information
type client struct {
	clientId     string
	clientSecret string
	redirectURIs []string
}

var demoClient = client{
	clientId:     "oauth-client-1",
	clientSecret: "oauth-client-secret-1",
	redirectURIs: []string{"http://localhost:9000/callback"},
}

//go:embed views
var clientFS embed.FS

func main() {
	engine := gin.Default()
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{"add": func(a, b int) int {
		return a + b
	}}).ParseFS(clientFS, "views/*.html"))
	engine.SetHTMLTemplate(tmpl)

	engine.GET("/", func(c *gin.Context) {
		viewData := gin.H{
			"accessToken": "NONE",
		}
		c.HTML(http.StatusOK, "index.html", viewData)
	})
	engine.GET("/authorize", authorize())
	engine.GET("/callback", callback())
	engine.GET("/fetch_resource", fetchResource())
	engine.Run(":9000")
}

func authorize() gin.HandlerFunc {

	/*
	 * Send the user to the authorization server
	 */

	return func(c *gin.Context) {}
}

func callback() gin.HandlerFunc {

	/*
	 * Parse the response from the authorization server and get a token
	 */

	return func(c *gin.Context) {}
}

func fetchResource() gin.HandlerFunc {

	/*
	 * Use the access token to call the resource server
	 */

	return func(c *gin.Context) {}
}

func buildUrl() {
}

func encodeClientCredentials() string {
	return ""
}
