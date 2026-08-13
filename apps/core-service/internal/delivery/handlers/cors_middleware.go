package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware sets permissive-but-controlled CORS headers. Origins are
// config-driven via CORS_ALLOWED_ORIGINS; when a request Origin is on the
// allowlist it is echoed back (required because Allow-Credentials is true).
// If allowed is empty it defaults to http://localhost:3000 (dev).
func CORSMiddleware(allowed []string) gin.HandlerFunc {
	if len(allowed) == 0 {
		allowed = []string{"http://localhost:3000"}
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		allowedSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if _, ok := allowedSet[origin]; ok {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// No matching Origin header: fall back to the first allowed origin
			// (preserves the original dev-only behaviour for non-browser clients).
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowed[0])
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-Role")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
