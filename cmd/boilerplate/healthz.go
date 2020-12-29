package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/viper"
)

// https://github.com/kelseyhightower/app-healthz/tree/master/healthz
// https://medium.com/metrosystemsro/kubernetes-readiness-liveliness-probes-best-practices-86c3cd9f0b4a

// TODO: write helper to turn custom errors into Error type using errors.Unwrap
// https://golangbot.com/custom-errors/

// Error represents a check error
type Error struct {
	Message     string            `json:"message"`     // error message
	Type        string            `json:"type"`        // type of check
	Description string            `json:"description"` // describe type of check for humans
	Metadata    map[string]string `json:"metadata"`    // check instance metadata (e.g. request id, trace id, ...)
}

// Response represents a healthz response
type Response struct {
	// healthz returns configuration
	Hostname string            `json:"hostname"`
	Metadata map[string]string `json:"metadata"`
	// healthz returns status
	Errors []Error `json:"errors"`
}

const (
	configGoroutineThreshold = "goroutine_threshold"
)

var (
	errorNumGoroutine = &Error{
		Message:     "readiness_NumGoroutine",
		Type:        "READINESS",
		Description: "runtime.NumGoroutine() exceeds threshold",
		Metadata: map[string]string{
			"GoroutineThreshold": "0",
			"NumGoroutine":       "0",
		},
	}
)

func init() {
	viper.SetDefault(configGoroutineThreshold, "1000")
}

// HealthzHandler is handling a ping request
func HealthzHandler(c *gin.Context) {
	log := hlog.FromRequest(c.Request)

	// check if the microservice can receive business traffic

	code := http.StatusOK
	host, err := os.Hostname()
	if err != nil {
		log.Error().Err(err).Msg("error getting hostname")
		c.Status(http.StatusInternalServerError)
		return
	}
	resp := &Response{Hostname: host}
	errors := make([]Error, 0)

	// this handler is returning readiness status unless it is used in a
	// liveness handler configured as follows in which case it's short
	// cutting and returning liveness status already here:
	//
	// livenessProbe:
	//   httpGet:
	//     path: /healthz
	//     port: 8080
	//     httpHeaders:
	//     - name: LIVENESS
	//       value: "1"
	//
	if hdr := c.Request.Header["LIVENESS"]; len(hdr) > 0 && hdr[0] != "" {
		c.JSON(code, resp)
		return
	}

	// if the status code for the ready probe is 4xx or 5xx then pod is marked
	// as unhealthy and HTTP traffic will no longer be redirected to it for
	// increasing reliability and uptime.

	// check number of go threads
	count := (uint)(runtime.NumGoroutine())
	threshold := viper.GetUint(configGoroutineThreshold)
	if count > threshold {
		err := *errorNumGoroutine
		err.Metadata["GoroutineThreshold"] = fmt.Sprint(threshold)
		err.Metadata["NumGoroutine"] = fmt.Sprint(count)
		errors = append(errors, err)
		log.Error().
			Str("GoroutineThreshold", err.Metadata["GoroutineThreshold"]).
			Str("NumGoroutine", err.Metadata["NumGoroutine"]).
			Msg(errorNumGoroutine.Description)
	}

	// ping database
	// add ping latency etc to metrics
	// on errors use logger to output details and `append(errors, Error{})`

	// the readiness handler must be application specific

	resp.Errors = errors
	if len(resp.Errors) > 0 {
		code = http.StatusInternalServerError
	}
	c.JSON(code, resp)
}

// type handler struct {
// 	hostname string
// 	metadata map[string]string
// }
//
// type Config struct {
// 	Hostname string
// }
//
// func HealthzHandler(c *Config) (gin.HandlerFunc, error) {
// 	// check configuration
// 	metadata := make(map[string]string)
// 	metadata["database_url"] = "test-addr"
//
// 	h := &handler{hostname: c.Hostname, metadata: metadata}
// 	return h, nil
// }
