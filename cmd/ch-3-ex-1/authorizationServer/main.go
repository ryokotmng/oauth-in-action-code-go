package main

import (
	"embed"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

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

var requests map[string]url.Values

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
	engine.GET("/authorize", authorize)
	engine.POST("/approve", approve)
	engine.POST("/token", token)
	engine.Run(":9001")
	fmt.Println("OAuth Authorization Server is listening at http://localhost:9000")
}

func authorize(c *gin.Context) {
	clientID := c.Request.URL.Query().Get("client_id")
	cl, ok := clients[clientID]
	if !ok {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Unknown client"})
		return
	}

	uri := c.Request.URL.Query().Get("redirect_uri")
	if !contains(cl.redirectURIs, uri) {
		fmt.Sprintf("Mismatched redirect URI, expected %s got %s", cl.redirectURIs, uri)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid redirect url"})
		return
	}
	rscope := strings.Split(c.Request.URL.Query().Get("scope"), " ")
	cscope := strings.Split(cl.scope, " ")
	if len(rscope) > len(cscope) {
		redirectURI := c.Request.URL.Query().Get("redirect_uri")
		url, err := url.Parse(redirectURI)
		if err != nil {
			return
		}
		q := url.Query()
		q.Set("error", "invalid_scope")
		url.RawQuery = q.Encode()
		c.Redirect(302, url.String())
		return
	}

	reqid := randomString(8)

	requests[reqid] = c.Request.URL.Query()

	viewData := gin.H{
		"client": cl, "reqid": reqid, "scope": rscope,
	}
	c.HTML(http.StatusOK, "approve.html", viewData)
}

func approve(c *gin.Context) {
}

func token(c *gin.Context) {
}

func contains(sl []string, str string) bool {
	for _, s := range sl {
		if s == str {
			return true
		}
	}
	return false
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
