package zerolog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	. "github.com/jblunck/bp/internal/zerolog"
)

func TestMiddleware(t *testing.T) {
	out := &bytes.Buffer{}
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, r := gin.CreateTestContext(resp)
	r.Use(Middleware(zerolog.New(out)))
	r.GET("/", func(c *gin.Context) {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("unknown error"))
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "514bbe5bb5251c92bd07a9846f4a1ab6")
	req.Header.Set("X-B3-TraceId", "80f198ee56343ba864fe8b2a57d3eff7")
	req.Header.Set("X-B3-ParentSpanId", "05e3ac9a4f6e3b90")
	req.Header.Set("X-B3-SpanId", "e457b5a2e4d86bd1")
	req.Header.Set("X-B3-Sampled", "1")
	c.Request = req
	r.ServeHTTP(resp, c.Request)

	var jsonLog struct {
		Method    string
		Path      string
		RequestID string `json:"request_id"`
		Span      struct {
			ID string
		}
		Trace struct {
			ID string
		}
	}
	json.Unmarshal(out.Bytes(), &jsonLog)

	assert.Equal(t, "GET", jsonLog.Method)
	assert.Equal(t, "/", jsonLog.Path)
	assert.Equal(t, "514bbe5bb5251c92bd07a9846f4a1ab6", jsonLog.RequestID)
	assert.Equal(t, "e457b5a2e4d86bd1", jsonLog.Span.ID)
	assert.Equal(t, "80f198ee56343ba864fe8b2a57d3eff7", jsonLog.Trace.ID)
}
