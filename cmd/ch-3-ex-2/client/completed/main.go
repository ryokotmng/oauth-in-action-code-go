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
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/ryokotmng/oauth-in-action-code-go/pkg"
)

const protectedResource = "http://localhost:9002/resource"

type tokenResponseBody struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

var (
	state  string
	scope  string
	client = oauth2.Config{
		ClientID:     "oauth-client-1",
		ClientSecret: "oauth-client-secret-1",
		RedirectURL:  "http://localhost:9000/callback",
		Scopes:       []string{"foo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:9001/authorize",
			TokenURL: "http://localhost:9001/token",
		},
	}
	refreshToken = "j2r3oj32r23rmasd98uhjrk2o3i"
	accessToken  = "987tghjkiu6trfghjuytrghj"
)

// move this file to "client" folder and uncomment this to use view files
// //go:embed views
var clientFS embed.FS

func main() {
	router := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/*.html"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/", func(c *gin.Context) {
		viewData := gin.H{"accessToken": accessToken, "scope": scope, "refreshToken": refreshToken}
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
		"scope":         strings.Join(client.Scopes, " "),
		"client_id":     client.ClientID,
		"redirect_uri":  client.RedirectURL,
		"state":         state,
	})

	fmt.Printf("redirect %s \n", authorizeUrl)
	c.Redirect(302, authorizeUrl)
}

func callback(c *gin.Context) {

	if c.Errors != nil {
		// it's an error response, act accordingly
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

		if resBody.RefreshToken != "" {
			refreshToken = resBody.RefreshToken
			fmt.Printf("Got refresh token: %s \n", refreshToken)
		}

		scope := resBody.Scope
		fmt.Printf("Got scope: %s", scope)

		c.HTML(http.StatusOK, "index.html", gin.H{"accessToken": accessToken, "scope": scope, "refreshToken": refreshToken})
		return
	}

	c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": fmt.Sprintf("Unable to fetch access token, server response: %v", tokRes.StatusCode)})
}

func fetchResource(c *gin.Context) {

	fmt.Printf("Making request with access token %s \n", accessToken)

	req, err := http.NewRequest("POST", protectedResource, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resource, err := http.DefaultClient.Do(req)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
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
	if refreshToken != "" {
		refreshAccessToken(c)
		return
	}
	c.HTML(resource.StatusCode, "error.html", gin.H{"error": resource.StatusCode})
}

func refreshAccessToken(c *gin.Context) {
	formData, err := json.Marshal(map[string]string{
		"grant_type":    "authorization_code",
		"refresh_token": refreshToken,
	})
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		return
	}
	req, err := http.NewRequest("POST", client.Endpoint.TokenURL, bytes.NewBuffer(formData))
	// you can use oauth2 package to get tokens like this: client.Exchange(c, code)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	fmt.Printf("Refreshing token %s \n", refreshToken)
	tokRes, err := http.DefaultClient.Do(req)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		return
	}
	if tokRes.StatusCode >= 200 && tokRes.StatusCode < 300 {
		bodyReader := tokRes.Body
		body, err := io.ReadAll(bodyReader)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		}
		var bodyData tokenResponseBody
		err = json.Unmarshal(body, &bodyData)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
			return
		}
		fmt.Printf("Got access token: %s \n", bodyData.AccessToken)
		if bodyData.RefreshToken != "" {
			refreshToken = bodyData.RefreshToken
			fmt.Printf("Got refresh token: %s \n", refreshToken)
		}
		scope = bodyData.Scope
		fmt.Printf("Got scope :%s \n", scope)

		// try again
		c.Redirect(302, "/fetch_resource")
		return
	}
	fmt.Println("No refresh token, asking the user to get a new access token")
	// tell the user to get a new access token
	refreshToken = ""
	c.HTML(tokRes.StatusCode, "error.html", gin.H{"error": "Unable to refresh token."})
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
