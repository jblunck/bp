package zerolog

import (
	"os"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

// Middleware returns a handler for gin
func Middleware() gin.HandlerFunc {
	subLog := log.With().
		Str("foo", "bar").
		Logger()

	return logger.SetLogger(logger.Config{
		Logger: &subLog,
		UTC:    true,
	})
}
