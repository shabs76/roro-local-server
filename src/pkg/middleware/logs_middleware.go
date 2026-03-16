package middleware

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var passwordRegex = regexp.MustCompile(`("password"\s*:\s*)"[^"]+"`)

// RequestLogger returns a middleware that logs request details.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read the Body
		var bodyBytes []byte
		// We avoid reading large file uploads into memory for logging
		contentType := c.GetHeader("Content-Type")
		isMultipart := strings.Contains(contentType, "multipart/form-data")

		if c.Request.Body != nil && !isMultipart {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restore the io.ReadCloser to its original state
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		} else if isMultipart {
			bodyBytes = []byte("[Multipart/Form-Data Content Omitted]")
		}

		// Process Request
		c.Next()

		// Skip logging for health checks if desired, but user didn't ask to skip.
		// Construct the log message
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		// Format Headers for display
		var headersBuffer bytes.Buffer
		for k, v := range c.Request.Header {
			fmt.Fprintf(&headersBuffer, "\n    %s: %s", k, strings.Join(v, ", "))
		}

		// Redact password from payload
		payloadStr := string(bodyBytes)
		payloadStr = passwordRegex.ReplaceAllString(payloadStr, `$1"*****"`)

		// Create a human-readable block
		// We write directly to gin.DefaultWriter to avoid JSON escaping by slog
		logMessage := fmt.Sprintf(`
--------------------------------------------------------------------------------
REQUEST REPORT
--------------------------------------------------------------------------------
Timestamp : %s
Method    : %s
Path      : %s
Client IP : %s
Latency   : %v
Status    : %d

HEADERS   :%s

PAYLOAD   :
%s
--------------------------------------------------------------------------------
`,
			start.Format(time.RFC3339),
			c.Request.Method,
			path,
			c.ClientIP(),
			time.Since(start),
			c.Writer.Status(),
			headersBuffer.String(),
			payloadStr,
		)

		// Log directly to the configured writer (file + stdout) to preserve formatting
		fmt.Fprint(gin.DefaultWriter, logMessage)
	}
}
