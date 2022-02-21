package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed views
var clientFS embed.FS

func main() {
	router := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/index.html"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/", func(c *gin.Context) {
		viewData := gin.H{
			"accessToken":  "NONE",
			"scope":        "NONE",
			"refreshToken": "NONE",
		}
		c.HTML(http.StatusOK, "index.html", viewData)
	})
	router.Run(":9000")
	fmt.Println("OAuth Client is listening at http://localhost:9000")
}
