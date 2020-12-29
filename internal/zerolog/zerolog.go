package zerolog

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

const (
	headerRequestID = "x-request-id"
	headerB3TraceID = "x-b3-traceid"
	headerB3SpanID  = "x-b3-spanid"
)

func init() {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:     os.Stderr,
				NoColor: false,
			},
		)
		log.Info().Msg("Running with human-friendly pretty logging enabled")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

// TracingHeaderHandler adds tracing headers to zerolog context
func TracingHeaderHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if val := r.Header.Get(headerB3TraceID); val != "" {
				log := zerolog.Ctx(r.Context())
				log.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Dict("trace", zerolog.Dict().Str("id", val))
				})
			}
			if val := r.Header.Get(headerB3SpanID); val != "" {
				log := zerolog.Ctx(r.Context())
				log.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Dict("span", zerolog.Dict().Str("id", val))
				})
			}
			next.ServeHTTP(w, r)
		})
	}
}

func handlerChain() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return hlog.CustomHeaderHandler("request_id", headerRequestID)(TracingHeaderHandler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})))
	}
}

// Middleware returns a handler for gin
func Middleware(log zerolog.Logger) gin.HandlerFunc {
	h := handlerChain()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// inject logger into context like hlog.NewHandler() would do
		l := log.With().Logger()
		c.Request = c.Request.WithContext(l.WithContext(c.Request.Context()))

		h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)

		end := time.Now()
		latency := end.Sub(start)
		msg := "Request"
		if len(c.Errors) > 0 {
			msg = c.Errors.String()
		}

		dl := hlog.FromRequest(c.Request).With().
			Int("status", c.Writer.Status()).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("ip", c.ClientIP()).
			Dur("latency", latency).
			Str("user-agent", c.Request.UserAgent()).
			Logger()

		switch {
		case c.Writer.Status() >= http.StatusInternalServerError:
			dl.Error().Msg(msg)
		case c.Writer.Status() >= http.StatusBadRequest:
			dl.Warn().Msg(msg)
		default:
			dl.Info().Msg(msg)
		}

		if c.Writer.Status() != http.StatusOK {
			c.AbortWithStatus(c.Writer.Status())
		}
	}
}
