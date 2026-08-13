package handlers

import (
	"bytes"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"rekberkuy/core-service/internal/domain"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// IdempotencyMiddleware intercepts the request path to prevent double execution
// (double-spending) based on the Idempotency-Key header. The first request is executed
// and its response is stored; a retried request with the same key gets an identical response.
func IdempotencyMiddleware(repo domain.IdempotencyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Read the Idempotency-Key from the header
		idempotencyKey := c.GetHeader("Idempotency-Key")

		// If the key is empty, the route does not require idempotency -> let it pass through
		if idempotencyKey == "" {
			c.Next()
			return
		}

		// 2. Build the initial draft record object (response still empty)
		record := &domain.IdempotencyRecord{
			ID:          idempotencyKey,
			RequestPath: c.Request.URL.Path,
		}

		// 3. Atomically check/lock (INSERT ... ON CONFLICT / SETNX)
		existingRecord, isNew, err := repo.CheckOrLock(c.Request.Context(), record)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Financial protection system error: " + err.Error()})
			c.Abort()
			return
		}

		// 4. Duplicate detected -> replay the cached response, UNLESS the original
		//    request is still in-flight (placeholder row, ResponseStatus == 0). In that
		//    race window, ask the client to retry shortly instead of returning a malformed
		//    empty response (and never double-execute the handler).
		if !isNew {
			if existingRecord.ResponseStatus == 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "a request with this Idempotency-Key is still processing; retry shortly"})
				c.Abort()
				return
			}
			c.Header("X-Cache-Idempotency", "true")
			c.Data(existingRecord.ResponseStatus, "application/json", existingRecord.ResponseBody)
			c.Abort()
			return
		}

		// 5. New request -> capture the response body to store after the handler finishes
		blw := &bodyLogWriter{body: bytes.NewBuffer(nil), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 6. Persist the original response (status + body) to the store. The next retried
		//    request with the same key will get an identical response -> prevents double-spending.
		if err := repo.SaveResponse(c.Request.Context(), idempotencyKey, blw.Status(), blw.body.Bytes()); err != nil {
			// Not a fatal failure for this request (response already sent to client),
			// but log it because retried requests will not dedup correctly.
			log.Printf("[IDEMPOTENCY] failed to save response key=%s: %v", idempotencyKey, err)
		}
	}
}
