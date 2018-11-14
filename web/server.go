package web

import (
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()


	//添加静态文件
	r.Use(static.Serve("/", static.LocalFile("./views", true)))


	//设置首页
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/index.html")
	})

	//其他页
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}