package otel

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// https://www.jaegertracing.io/docs/1.19/client-features/
const (
	endpointVar    = "JAEGER_ENDPOINT"
	serviceNameVar = "JAEGER_SERVICE_NAME"
)

func getServiceName() string {
	serviceName, ok := os.LookupEnv(serviceNameVar)
	if !ok {
		serviceName = path.Base(os.Args[0])
	}
	return serviceName
}

type logErrorHandler struct {
	l zerolog.Logger
}

func (h logErrorHandler) Handle(e error) {
	h.l.Err(e).Msgf("OpenTelemetry: %s", e.Error())
}

func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn().Err(err).Msg("hostname lookup failure; using \"localhost\"")
		return "localhost"
	}
	return hostname
}

// InitTracer initializes tracing for an environment with Jaeger installed
func InitTracer(appInfo map[string]string) func() {
	endpoint, ok := os.LookupEnv(endpointVar)
	if !ok {
		if host, ok := os.LookupEnv("JAEGER_COLLECTOR_SERVICE_HOST"); ok {
			port := os.Getenv("JAEGER_COLLECTOR_SERVICE_PORT_JAEGER_COLLECTOR_HTTP")
			endpoint = fmt.Sprintf("http://%s:%s/api/traces", host, port)
		}
	}

	if endpoint == "" {
		log.Info().Msg("no jaeger collector found")
		return func() {}
	}

	// Register the B3 propagator globally.
	b3 := b3.B3{}
	otel.SetTextMapPropagator(b3)

	l := log.With().Dict("otel", zerolog.Dict().
		Str("version", otel.Version()).
		Str("endpoint", endpoint).
		Str("servicename", getServiceName())).
		Logger()
	eh := logErrorHandler{l: l}
	otel.SetErrorHandler(eh)

	tags := []label.KeyValue{
		label.String("cmd", path.Base(os.Args[0])),
		label.String("args", strings.Join(os.Args[1:], " ")),
		label.String("hostname", hostname()),
	}
	// append app information
	for k, v := range appInfo {
		tags = append(tags, label.String(k, v))
	}
	// append k8s metadata from downward API
	envs := []string{"METADATA_NAME", "METADATA_NAMESPACE", "METADATA_UID"}
	for _, env := range envs {
		if val, ok := os.LookupEnv(env); ok {
			tags = append(tags, label.String(env, val))
		}
	}

	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(endpoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: getServiceName(),
			Tags:        tags,
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	l.Err(err).Msg("Jaeger export pipeline")
	return flush
}
