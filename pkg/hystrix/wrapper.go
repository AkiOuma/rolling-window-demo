package hystrix

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Wrapper(
	size,
	reqThreshold int,
	failedThreshold float64,
	duration time.Duration,
) gin.HandlerFunc {
	r := NewRollingWindow(size, reqThreshold, failedThreshold, duration)
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
