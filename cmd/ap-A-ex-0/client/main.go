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
	engine := gin.Default()
	tmpl := template.Must(template.ParseFS(clientFS, "views/index.html"))
	engine.SetHTMLTemplate(tmpl)

	engine.GET("/", func(c *gin.Context) {
		viewData := gin.H{
			"accessToken":  "NONE",
			"scope":        "NONE",
			"refreshToken": "NONE",
		}
		c.HTML(http.StatusOK, "index.html", viewData)
	})
	engine.Run(":9000")
	fmt.Println("OAuth Client is listening at http://localhost:9000")
}
