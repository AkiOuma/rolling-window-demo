package server

import (
	"hystrix-demo/pkg/hystrix"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Wrapper() gin.HandlerFunc {
	r := hystrix.NewRollingWindow(10, 10, 0, time.Second*5)
	r.Launch()
	r.Monitor()
	r.ShowStatus()
	return func(c *gin.Context) {
		if r.Broken() {
			c.String(http.StatusInternalServerError, "reject by hystrix")
			c.Abort()
			return
		}
		c.Next()
		if c.Writer.Status() != http.StatusOK {
			r.RecordReqResult(false)
		} else {
			r.RecordReqResult(true)
		}
	}
}

func NewUpStreamServer() *gin.Engine {
	app := gin.Default()
	app.GET("/api/up/v1", Wrapper(), upHandler)
	return app
}

func upHandler(c *gin.Context) {
	res, err := http.Get("http://localhost:8000/api/down/v1")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err != nil {
		c.String(res.StatusCode, string(data))
		return
	}
	c.String(res.StatusCode, string(data))
}
