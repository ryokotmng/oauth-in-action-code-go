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
	"golang.org/x/oauth2"

	"github.com/ryokotmng/oauth-in-action-code-go/pkg"
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
	state = pkg.RandomString(32)
	authorizeUrl := buildUrl(client.Endpoint.AuthURL, &map[string]string{
		"response_type": "code",
		"client_id":     client.ClientID,
		"redirect_uri":  client.RedirectURL,
		"state":         state,
	})

	c.Writer.Status()
	c.Redirect(302, authorizeUrl)
}

func callback(c *gin.Context) {

	// it's an error response, act accordingly
	if c.Errors != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": c.Errors})
		return
	}

	if s := c.Query("state"); s != state {
		fmt.Printf("State DOES NOT MATCH: expected %s got %s \n", state, s)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "State value did not match"})
		return
	}

	code := c.Query("code")

	formData, err := json.Marshal(map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": client.RedirectURL,
	})
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest(http.MethodPost, client.Endpoint.TokenURL, bytes.NewBuffer(formData))
	// you can use oauth2 package to get tokens like this: client.Exchange(c, code)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+encodeClientCredentials(client.ClientID, client.ClientSecret))
	tokRes, err := http.DefaultClient.Do(req)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
	}

	fmt.Printf("Requesting access token for code %s \n", code)
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
		accessToken = resBody.AccessToken
		fmt.Printf("Got access token: %s \n", accessToken)

		c.HTML(http.StatusOK, "index.html", gin.H{"accessToken": accessToken, "scope": scope})
		return
	}

	c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": fmt.Sprintf("Unable to fetch access token, server response: %v", tokRes.StatusCode)})
}

func fetchResource(c *gin.Context) {
	if accessToken == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Missing Access Token"})
		return
	}

	fmt.Printf("Making request with access token %s \n", accessToken)

	req, err := http.NewRequest("POST", protectedResource, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resource, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resource.StatusCode >= 200 && resource.StatusCode < 300 {
		bodyReader := resource.Body
		body, err := io.ReadAll(bodyReader)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		}
		var bodyData interface{}
		json.Unmarshal(body, &bodyData)
		c.HTML(http.StatusOK, "data.html", gin.H{"resource": bodyData})
		return
	}
	accessToken = ""
	c.HTML(resource.StatusCode, "error.html", gin.H{"error": resource.StatusCode})
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
