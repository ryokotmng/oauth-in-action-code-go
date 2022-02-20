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
		c.HTML(http.StatusOK, "index.html", nil)
	})
	engine.POST("/resource", resource())
	engine.Run(":9002")
	fmt.Println("OAuth Resource Server is listening at http://localhost:9002")
}

func resource() gin.HandlerFunc {
	return func(c *gin.Context) {}
}
