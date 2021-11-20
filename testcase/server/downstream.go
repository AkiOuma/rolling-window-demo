package server

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func NewDownStreamServer(successRate float64) *gin.Engine {
	if successRate > 1 || successRate < 0 {
		panic("invalid rate")
	}
	app := gin.Default()
	app.GET("/api/down/v1", func(c *gin.Context) {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
		if !rejectOrNot(successRate) {
			c.String(http.StatusInternalServerError, "reject from downstream")
			return
		}
		c.String(http.StatusOK, "approve from downstream")
	})
	return app
}

func rejectOrNot(successRate float64) bool {
	return rand.Float64() < successRate
}
