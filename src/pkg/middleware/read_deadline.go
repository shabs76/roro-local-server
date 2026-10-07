package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// BodyReadDeadline limits how long the request body of a POST/PUT/PATCH/DELETE may
// take to arrive. The server sets no global ReadTimeout: that would also cut slow
// photo uploads, and it cancels long-lived GET streams (SSE) because net/http aborts
// the request context when the connection read deadline passes. GET and HEAD requests
// are left without a deadline for the same reason.
func BodyReadDeadline(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			setReadDeadline(c, d)
		}
		c.Next()
	}
}

func setReadDeadline(c *gin.Context, d time.Duration) {
	rc := http.NewResponseController(c.Writer)
	if err := rc.SetReadDeadline(time.Now().Add(d)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		slog.Warn("Failed to set request read deadline", "path", c.Request.URL.Path, "error", err)
	}
}
