package middleware

import (
	"net/http"
	"single-agent/observability"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type TracingMiddleware struct {
	telemetry *observability.Telemetry
}

func NewTracingMiddleware(telemetry *observability.Telemetry) *TracingMiddleware {
	return &TracingMiddleware{
		telemetry: telemetry,
	}
}

func (t *TracingMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if t.telemetry == nil || t.telemetry.Tracer == nil {
			// Tracing not enable, pass through
			c.Next()
			return
		}

		// Extract trace context from incoming request headers (W3C Trace Context)
		ctx := c.Request.Context()
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		// Start a new span for this http
		ctx, span := t.telemetry.Tracer.Start(
			ctx, "http.request",
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.Path),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Pass the new context to the gin request
		c.Request = c.Request.WithContext(ctx)

		// Call the next handler
		c.Next()

		// Record span attributes based on response
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// Set span status based on HTTP status code
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, http.StatusText(c.Writer.Status()))
		} else {
			span.SetStatus(codes.Ok, "Request completed successfully")
		}
	}
}
