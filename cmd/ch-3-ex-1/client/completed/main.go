package main

import (
	"embed"
	"encoding/base64"
	"html/template"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// authorization server information
const (
	// authServer
	authorizationEndpoint = "http://localhost:9001/authorize"
	tokenEndpoint         = "http://localhost:9001/token"

	protectedResource = "http://localhost:9002/resource"
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

var (
	state       string
	accessToken string
	scope       string
)

// move this file to "client" folder and uncomment this to use view files
// //go:embed views
var clientFS embed.FS

func main() {
	router := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/*.html"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/", func(c *gin.Context) {
		viewData := gin.H{"accessToken": accessToken, "scope": scope}
		c.HTML(http.StatusOK, "index.html", viewData)
	})
	router.GET("/authorize", authorize)
	router.GET("/callback", callback)
	router.GET("/fetch_resource", fetchResource)
	router.Run(":9000")
}

func authorize(c *gin.Context) {

	/*
	 * Send the user to the authorization server
	 */

}

func callback(c *gin.Context) {

	/*
	 * Parse the response from the authorization server and get a token
	 */

}

func fetchResource(c *gin.Context) {

	/*
	 * Use the access token to call the resource server
	 */

}

func buildUrl(base string, options, hash map[string]string) *url.URL {
	newUrl, err := url.Parse(base)
	if err != nil {
		return nil
	}

	q := newUrl.Query()
	for k, v := range options {
		q.Set(k, v)
	}
	newUrl.RawQuery = q.Encode()
	return newUrl
}

func encodeClientCredentials(clientId, clientSecret string) *base64.Encoding {
	return base64.NewEncoding(url.QueryEscape(clientId) + ":" + url.QueryEscape(clientSecret))
}
