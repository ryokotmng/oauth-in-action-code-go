package pkg

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type key int

const (
	accessTokenKey key = iota
)

// SetAccessTokenKeyToContext adds specified key and value, then returns http.Request
func SetAccessTokenKeyToContext(c gin.Context, token string) *http.Request {
	return c.Request.WithContext(context.WithValue(c.Request.Context(), accessTokenKey, token))
}

// GetAccessTokenFromContext returns access token value and bool which shows if the value exists or not
func GetAccessTokenFromContext(c *gin.Context) (accessToken string, ok bool) {
	if c == nil {
		return "", false
	}
	accessToken, ok = c.Request.Context().Value(accessToken).(string)
	return accessToken, ok
}
