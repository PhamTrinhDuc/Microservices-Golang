package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	// Request metrics
	RequestCount    metric.Int64Counter       // đếm số request
	RequestDuration metric.Float64Histogram   // đo thời gian request
	ActiveRequests  metric.Int64UpDownCounter // đếm số request đang hoạt động

	// Tool execution metrics
	ToolExecutionCount      metric.Int64Counter     // đếm số lần tool được gọi
	ToolExecutionDuration   metric.Float64Histogram // đo thời gian tool được gọi
	ToolExecutionErrorCount metric.Int64Counter     // đếm số lần tool bị lỗi

	// Database metrics
	DBQueryDuration        metric.Float64Histogram   // đo thời gian query database
	DBQueryCount           metric.Int64Counter       // đếm số lần query database
	DBConnectionPoolActive metric.Int64UpDownCounter // số kết nối đang hoạt động
	DBConnectionPoolIdle   metric.Int64UpDownCounter // số kết nối rảnh

	// Search metrics
	SearchResultCount   metric.Int64Histogram   // số kết quả tìm kiếm
	HybridSearchLatency metric.Float64Histogram // thời gian tìm kiếm hybrid
	HybridSearchScore   metric.Float64Histogram // điểm số tìm kiếm hybrid
	VectorSearchLatency metric.Float64Histogram // thời gian tìm kiếm vector
	BM25SearchLatency   metric.Float64Histogram // thời gian tìm kiếm BM25

	// Document metrics
	DocumentRetrievedCount metric.Int64Counter // số tài liệu được truy xuất

	// Err metrics
	ErrorCount metric.Int64Counter // đếm số lỗi
}

func NewMetrics(meter metric.Meter) (*Metrics, error) {
	m := &Metrics{}
	var err error

	// 1. Request metrics
	m.RequestCount, err = meter.Int64Counter(
		"mcp.request.count",
		metric.WithDescription("Tổng số lượng requests tới MCP Server"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init request count metric: %w", err)
	}

	m.RequestDuration, err = meter.Float64Histogram(
		"mcp.request.duration",
		metric.WithDescription("Thời gian request MCP (miliseconds)"),
		metric.WithUnit("ms"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to init request duration metric: %w", err)
	}

	m.ActiveRequests, err = meter.Int64UpDownCounter(
		"mcp.request.active",
		metric.WithDescription("Số request đang hoạt động"),
		metric.WithUnit("{request}"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to init request active metric: %w", err)
	}

	// 2. Tool metrics
	m.ToolExecutionCount, err = meter.Int64Counter(
		"mcp.tool.execution",
		metric.WithDescription("Tổng số tool thực hiện"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init tool execution metrics: %w", err)
	}

	m.ToolExecutionDuration, err = meter.Float64Histogram(
		"mcp.tool.duration",
		metric.WithDescription("Thời gian tool thực hiện (miliseconds)"),
		metric.WithUnit("ms"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to init tool duration metric: %w", err)
	}

	m.ToolExecutionErrorCount, err = meter.Int64Counter(
		"mcp.tool.errors",
		metric.WithDescription("Tổng số lỗi khi thực thi tool"),
		metric.WithUnit("{error}"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create tool count error metrics: %w", err)
	}

	// 3. Database metrics
	m.DBQueryCount, err = meter.Int64Counter(
		"mcp.db.query.count",
		metric.WithDescription("Tổng số request tới VDB"),
		metric.WithUnit("{query}"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to init db query count metrics: %w", err)
	}

	m.DBQueryDuration, err = meter.Float64Histogram(
		"mcp.db.query.duration",
		metric.WithDescription("Thời gian VDB thực thi query"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init db query duration metrics: %w", err)
	}

	m.DBConnectionPoolActive, err = meter.Int64UpDownCounter(
		"mcp.db.con_pool.activate",
		metric.WithDescription("Các connection pool đang hoạt động"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init db connection pool activate metrics: %w", err)
	}

	m.DBConnectionPoolIdle, err = meter.Int64UpDownCounter(
		"mcp.db.con_pool.idle",
		metric.WithDescription("Các connection đang rảnh"),
		metric.WithUnit("{connection}"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to init db connecton pool idle: %w", err)
	}

	// 4. Search metrics
	m.SearchResultCount, err = meter.Int64Histogram(
		"mcp.search.results",
		metric.WithDescription("Số lượng kết quả search trả về"),
		metric.WithUnit("{result}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create search result count metric: %w", err)
	}

	m.HybridSearchScore, err = meter.Float64Histogram(
		"mcp.search.hybrid_score",
		metric.WithDescription("Điểm số hybrid search"),
		metric.WithUnit("{score}"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create hybrid searcg score metric: %w", err)
	}

	// 5. Document metrics
	m.DocumentRetrievedCount, err = meter.Int64Counter(
		"mcp.document.count",
		metric.WithDescription("Số document retriever được"),
		metric.WithUnit("{document}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create document retrieved count metric: %w", err)
	}

	m.ErrorCount, err = meter.Int64Counter(
		"mcp.error.count",
		metric.WithDescription("Số lượng lỗi thực thi MCP"),
		metric.WithUnit("{error}"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to creat error count metric: %w", err)
	}

	return m, nil
}

// RecordRequest records metrics for an MCP request
func (m *Metrics) RecordRequest(ctx context.Context, method string, status string, durationMs float64) {
	atrrs := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("status", status),
	)

	m.RequestCount.Add(ctx, 1, atrrs)
	m.RequestDuration.Record(ctx, durationMs, atrrs)
}

// RecordError records an error occurrence
func (m *Metrics) RecordError(ctx context.Context, errorType string, operation string) {
	attrs := metric.WithAttributes(
		attribute.String("error.type", errorType),
		attribute.String("operation", operation),
	)

	m.ErrorCount.Add(ctx, 1, attrs)
}

func (m *Metrics) RecordToolExecution(ctx context.Context, toolName string, status string, durationMs float64) {
	atrrs := metric.WithAttributes(
		attribute.String("tool.name", toolName),
		attribute.String("status", status),
	)

	m.RequestCount.Add(ctx, 1, atrrs)
	m.RequestDuration.Record(ctx, durationMs, atrrs)
}
