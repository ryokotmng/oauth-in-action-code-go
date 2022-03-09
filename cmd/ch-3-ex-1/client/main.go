package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/ryokotmng/oauth-in-action-code-go/pkg"
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

type tokenResponseBody struct {
	accessToken string `json:"access_token"`
}

var (
	state       string
	accessToken string
	scope       string
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
	state := pkg.RandomString(32)
	authorizeUrl := buildUrl(authorizationEndpoint, &map[string]string{
		"response_type": "code",
		"client_id":     demoClient.clientId,
		"redirect_uri":  demoClient.redirectURIs[0],
		"state":         state,
	}, nil)

	c.Writer.Status()
	c.Redirect(302, authorizeUrl)
}

func callback(c *gin.Context) {

	// it's an error response, act accordingly
	if c.Errors != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": c.Errors})
		return
	}

	if s := pkg.GetStateFromContext(c.Request.Context()); s != state {
		fmt.Printf("State DOES NOT MATCH: expected %s got %s", state, s)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "State value did not match"})
		return
	}

	code := pkg.GetCodeFromContext(c)

	formData, err := json.Marshal(map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": demoClient.redirectURIs[0],
	})
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/x-www-form-urlencoded")
	c.Header("Authorization", "Basic "+encodeClientCredentials(demoClient.clientId, demoClient.clientSecret))

	tokRes, err := http.Post(tokenEndpoint, "application/json", bytes.NewBuffer(formData))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
	}

	fmt.Printf("Requesting access token for code %s", code)
	if tokRes.StatusCode >= 200 && tokRes.StatusCode < 300 {
		body, err := io.ReadAll(tokRes.Body)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		}
		defer func() {
			err := tokRes.Body.Close()
			if err != nil {
				c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
			}
		}()
		var resBody tokenResponseBody
		err = json.Unmarshal(body, &resBody)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": fmt.Sprintf("Unable to fetch access token, err: %v", err.Error())})
			return
		}
		fmt.Printf("Got access token: %s", resBody.accessToken)

		c.HTML(http.StatusOK, "index.html", gin.H{"accessToken": accessToken, "scope": scope})
		return
	}

	c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": fmt.Sprintf("Unable to fetch access token, server response: %v", tokRes.StatusCode)})
}

func fetchResource(c *gin.Context) {

}

func buildUrl(base string, options, hash *map[string]string) string {
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
