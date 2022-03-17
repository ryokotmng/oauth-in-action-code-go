package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"html/template"
	"io"
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
	ClientId     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scope        string   `json:"scope"`
}

var clients = map[string]*client{
	"oauth-client-1": {
		ClientId:     "oauth-client-1",
		ClientSecret: "oauth-client-secret-1",
		RedirectURIs: []string{"http://localhost:9000/callback"},
		Scope:        "foo bar",
	},
}

type approveReq struct {
	authorizationEndPointRequest url.Values
	scope                        []string
	user                         string
}

var codes map[string]*approveReq

var requests map[string]url.Values

type tokenRequestBody struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RefreshToken string `json:"refresh_token"`
}

type tokenResponseBody struct {
	accessToken  string `json:"access_token"`
	tokenType    string `json:"token_type"`
	refreshToken string `json:"refresh_token"`
}

//go:embed views
var clientFS embed.FS

func main() {
	router := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/*.html"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"clients": clients, "authServer": "NONE"})
	})
	router.GET("/authorize", authorize)
	router.POST("/approve", approve)
	router.POST("/token", token)
	router.Run(":9001")
	fmt.Println("OAuth Authorization Server is listening at http://localhost:9000")
}

func authorize(c *gin.Context) {
	clientID := c.Request.URL.Query().Get("client_id")
	cl, ok := clients[clientID]
	if !ok {
		fmt.Printf("Unknown client %s \n", clientID)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Unknown client"})
		return
	}

	uri := c.Request.URL.Query().Get("redirect_uri")
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

	c.HTML(http.StatusOK, "approve.html", gin.H{"client": cl, "reqid": reqid, "scope": cscope})
}

func approve(c *gin.Context) {
	c.Request.ParseForm()
	reqid := c.Request.Form.Get("reqid")
	query := requests[reqid]
	delete(requests, reqid)
	if query == nil {
		// there was no matching saved request, this is an error
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "No matching authorization request"})
		return
	}

	if c.Request.Form.Get("approve") == "Approve" {
		if query.Get("response_type") == "code" {
			// user approved access
			code := pkg.RandomString(8)

			user := c.Request.Form.Get("user")

			var scope []string
			for k, v := range c.Request.Form {
				if strings.HasPrefix(k, "scope_") {
					scope = append(scope, strings.Replace(k, "scope_", "", 1))
				}
				fmt.Println(v)
			}
			client := clients[query.Get("client_id")]
			cscope := strings.Split(client.Scope, " ")
			if len(scope) > len(cscope) {
				// client asked for a scope it couldn't have
				urlParsed, err := url.Parse(query.Get("redirect_uri"))
				if err != nil {
					c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": c.Errors})
				}
				urlParsed.Query().Add("error", "invalid_scope")
				c.Redirect(302, urlParsed.String())
				return
			}
			codes = map[string]*approveReq{}
			codes[code] = &approveReq{query, scope, user}

			urlParsed, err := url.Parse(query.Get("redirect_uri"))
			if err != nil {
				c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": c.Errors})
			}
			q := urlParsed.Query()
			q.Add("code", code)
			q.Add("state", query.Get("state"))
			urlParsed.RawQuery = q.Encode()
			c.Redirect(302, urlParsed.String())
			return
		}
		// we got a response type we don't understand
		urlParsed, err := url.Parse(query.Get("redirect_uri"))
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": c.Errors})
		}
		urlParsed.Query().Add("error", "unsupported_response_type")
		return
	}
	// user denied access
	urlParsed, err := url.Parse(query.Get("redirect_uri"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": c.Errors})
	}
	c.Redirect(302, urlParsed.String())
}

func token(c *gin.Context) {
	auth := c.Request.Header.Get("authorization")
	var clientID string
	var clientSecret string
	if auth != "" {
		// check the auth header
		clientCredentials := strings.Split(strings.Replace(auth, "Basic ", "", 1), ":")

		clientID = clientCredentials[0]
		clientSecret = clientCredentials[1]
	}

	// otherwise, check the post body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
	}
	defer func() {
		err := c.Request.Body.Close()
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		}
	}()
	var reqBody tokenRequestBody
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": fmt.Sprintf("Unable to fetch client id, err: %v", err.Error())})
		return
	}
	if reqBody.ClientID != "" {
		if clientID != "" {
			// if we've already seen the client's credentials in the authorization header, this is an error
			fmt.Println("Client attempted to authenticate with multiple methods")
			c.JSON(401, gin.H{"error": "invalid_client"})
			return
		}

		clientID = reqBody.ClientID
		clientSecret = reqBody.ClientSecret
	}

	client := clients[clientID]
	if client == nil {
		fmt.Printf("Unknown client %s \n", client)
		c.JSON(401, gin.H{"error": "invalid_client"})
		return
	}

	if client.ClientSecret != clientSecret {
		fmt.Printf("Mismatched client secret, expected %s got %s \n", client.ClientSecret, reqBody.ClientSecret)
		c.JSON(401, gin.H{"error": "invalid_client"})
		return
	}

	redisClient := pkg.NewRedisClient()

	if reqBody.GrantType == "authorization_code" {

		code := codes[reqBody.Code]

		if code != nil {
			delete(codes, reqBody.Code) // burn our Code, it's been used
			if code.authorizationEndPointRequest.Get("client_id") == clientID {

				accessToken := pkg.RandomString(32)

				var cscope string
				if code.scope != nil {
					cscope = strings.Join(code.scope, " ")
				}

				record, err := json.Marshal(pkg.TokenRecord{ClientID: clientID})
				if err != nil {
					fmt.Printf("Failed to register access token. err: %s \n", err.Error())
				}
				err = redisClient.Set(c, accessToken, record, 0).Err()
				if err != nil {
					fmt.Printf("Failed to register access token. err: %s \n", err.Error())
				}

				fmt.Printf("Issuing access token %s \n", accessToken)
				fmt.Printf("with scope %s \n", cscope)

				tokenResponse := map[string]string{"access_token": accessToken, "token_type": "Bearer", "scope": cscope}

				c.JSON(200, tokenResponse)
				fmt.Printf("Issued tokens for code %s \n", reqBody.Code)

				return
			}
			c.JSON(400, gin.H{"error": "invalid_grant"})
			return
		}
		fmt.Printf("Unknown Code, %s \n", reqBody.Code)
		c.JSON(400, gin.H{"error": "invalid_grant"})
		return
	} else if reqBody.GrantType == "refresh_token" {
		cID, err := redisClient.Get(c, "refresh_token"+reqBody.RefreshToken).Result()
		if err != redis.Nil {
			if cID != clientID {
				fmt.Printf("Invalid client using a refresh token, expected %s got %s \n", cID, clientID)
				redisClient.Del(c, reqBody.RefreshToken)
				c.HTML(http.StatusBadRequest, "error.html", nil)
			}
			fmt.Printf("We found a matching refresh token: %s", reqBody.RefreshToken)
			accessToken := pkg.RandomString(32)
			tokenResponse, err := json.Marshal(tokenResponseBody{accessToken, "Bearer", reqBody.RefreshToken})
			if err != nil {
				fmt.Printf("Error %s \n", err.Error())
				return
			}
			fmt.Printf("Issuing access token %s for refresh token %s \n", accessToken, reqBody.RefreshToken)
			c.JSON(200, tokenResponse)
		} else {
			fmt.Println("No matching token was found.")
			c.HTML(http.StatusUnauthorized, "error.html", nil)
		}
	}
	fmt.Printf("Unknown grant type %s \n", reqBody.GrantType)
	c.JSON(400, gin.H{"error": "unsupported_grant_type"})
}
