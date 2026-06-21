package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	// semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	SamplingRate   float64
	EnableTracing  bool
	EnableMetrics  bool
}

type Telemetry struct {
	TraceProvider *sdktrace.TracerProvider
	MeterProvider *metric.MeterProvider
	Tracer        trace.Tracer
	Metrics       *Metrics
	config        Config
}

// NewTelemetry initializes OpenTelemetry with tracing and metrics
func NewTelemetry(ctx context.Context, cfg Config) (*Telemetry, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			// semconv.SchemaURL,
			"",
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentName(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	t := &Telemetry{config: cfg}

	// Init tracing
	if cfg.EnableTracing {
		if err := t.initTracing(ctx, res); err != nil {
			return nil, fmt.Errorf("failed to init tracing: %w", err)
		}
	}
	if cfg.EnableMetrics {
		if err := t.initMetrics(res); err != nil {
			return nil, fmt.Errorf("failed to init metrics: %w", err)
		}
	}

	return t, nil
}

func (t *Telemetry) initTracing(ctx context.Context, resource *resource.Resource) error {
	// 1. Khởi tạo exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(t.config.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // Use insecure for local development
	)
	if err != nil {
		return fmt.Errorf("failed to create OLTP exporter: %w", err)
	}

	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(t.config.SamplingRate),
	)

	// 2. Khởi tạo Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512)),
		sdktrace.WithResource(resource),
	)

	// 3. Đưa vào Global - Đây là bước quan trọng nhất để các thư viện khác hoạt động
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	t.TraceProvider = tp
	t.Tracer = tp.Tracer(t.config.ServiceName)
	return nil
}

func (t *Telemetry) initMetrics(resource *resource.Resource) error {
	// 1. Khởi tạo exporter
	exporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to create Promethues exporter: %w", err)
	}
	// 2. Khởi tạo provider
	mp := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(exporter),
	)

	// 3. Đưa vào Global - Đây là bước quan trọng nhất để các thư viện khác hoạt động
	otel.SetMeterProvider(mp)

	t.MeterProvider = mp
	meter := mp.Meter(t.config.ServiceName)
	metrics, err := NewMetrics(meter)
	if err != nil {
		return fmt.Errorf("failed to create metric: %w", err)
	}
	t.Metrics = metrics
	return nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	var err error

	if t.TraceProvider != nil {
		if shutdownErr := t.TraceProvider.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("failed to shutdown tracer provider: %w", shutdownErr)
		}
	}

	if t.MeterProvider != nil {
		if shutdownErr := t.MeterProvider.Shutdown(ctx); shutdownErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; failed to shutdown meter provider: %w", err, shutdownErr)
			} else {
				err = fmt.Errorf("failed to shutdown meter provider: %w", shutdownErr)
			}
		}
	}

	return err
}
