package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ryokotmng/oauth-in-action-code-go/pkg"
)

const (
	authorizationEndpoint = "http://localhost:9001/authorize"
	tokenEndpoint         = "http://localhost:9001/token"
)

type client struct {
	ClientId     string
	ClientSecret string
	RedirectURIs []string
	Scope        string
}

var clients = map[string]client{
	"oauth-client-1": {
		ClientId:     "oauth-client-1",
		ClientSecret: "oauth-client-secret-1",
		RedirectURIs: []string{"http://localhost:9000/callback"},
		Scope:        "foo bar",
	},
}

var codes []string

var requests map[string]url.Values

//go:embed views
var clientFS embed.FS

func main() {
	router := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/*.html"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/", func(c *gin.Context) {
		viewData := gin.H{
			"clients":    clients,
			"authServer": "NONE",
		}
		c.HTML(http.StatusOK, "index.html", viewData)
	})
	router.GET("/authorize", authorize)
	router.POST("/approve", approve)
	router.POST("/token", token)
	router.Run(":9001")
	fmt.Println("OAuth Authorization Server is listening at http://localhost:9000")
}

func authorize(c *gin.Context) {
	clientID := c.Request.URL.Query().Get("clientId")
	cl, ok := clients[clientID]
	if !ok {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Unknown client"})
		return
	}

	uri := c.Request.URL.Query().Get("redirectUri")
	if !pkg.Contains(cl.RedirectURIs, uri) {
		fmt.Sprintf("Mismatched redirect URI, expected %s got %s", cl.RedirectURIs, uri)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid redirect url"})
		return
	}
	rscope := strings.Split(c.Request.URL.Query().Get("scope"), " ")
	cscope := strings.Split(cl.Scope, " ")
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

	reqid := pkg.RandomString(8)

	requests = map[string]url.Values{}
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
