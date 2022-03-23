package main

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const protectedResource = "http://localhost:9002/resource"

type tokenResponseBody struct {
	AccessToken string `json:"access_token"`
}

var (
	state       string
	accessToken string
	scope       string
	client      = &oauth2.Config{
		ClientID:     "oauth-client-1",
		ClientSecret: "oauth-client-secret-1",
		RedirectURL:  "http://localhost:9000/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:9001/authorize",
			TokenURL: "http://localhost:9001/token",
		},
	}
)

//go:embed views
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

func buildUrl(base string, options *map[string]string) string {
	newUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}

	q := newUrl.Query()
	for k, v := range *options {
		q.Set(k, v)
	}
	newUrl.RawQuery = q.Encode()
	return newUrl.String()
}

func encodeClientCredentials(clientId, clientSecret string) string {
	return url.QueryEscape(clientId) + ":" + url.QueryEscape(clientSecret)
}
