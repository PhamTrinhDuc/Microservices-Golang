package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestNewMetrics(t *testing.T) {
	// 1. Tạo Meter giả (Noop - No Operation)
	meterProvider := noop.NewMeterProvider()
	meter := meterProvider.Meter("test-meter")

	// 2. Test hàm khởi tạo
	m, err := NewMetrics(meter)

	// 3. Kiểm tra xem có lỗi không
	assert.NoError(t, err)
	assert.NotNil(t, m)

	// 4. Kiểm tra xem các biến metrics đã được khởi tạo chưa (không phải nil)
	assert.NotNil(t, m.RequestCount)
	assert.NotNil(t, m.RequestDuration)
	assert.NotNil(t, m.ErrorCount)
}

func TestRecordingMetrics(t *testing.T) {
	meterProvider := noop.NewMeterProvider()
	meter := meterProvider.Meter("test-meter")

	m, _ := NewMetrics(meter)

	ctx := context.Background()

	assert.NotPanics(t, func() {
		m.RequestCount.Add(ctx, 1)
		m.RequestDuration.Record(ctx, 150.5)
		m.ActiveRequests.Add(ctx, 1)
	})

	assert.NotPanics(t, func() {
		m.ToolExecutionCount.Add(ctx, 1)
		m.ToolExecutionDuration.Record(ctx, 1.0)
		m.ToolExecutionErrorCount.Add(ctx, 5)
	})

	assert.NotPanics(t, func() {
		m.DBQueryCount.Add(ctx, 1)
		m.DBQueryDuration.Record(ctx, 0.5)
		m.DBConnectionPoolActive.Add(ctx, 5)
		m.DBConnectionPoolIdle.Add(ctx, 4)
	})

	assert.NotPanics(t, func() {
		m.SearchResultCount.Record(ctx, 10)
		m.HybridSearchScore.Record(ctx, 0.7)
	})

	assert.NotPanics(t, func() {
		m.DocumentRetrievedCount.Add(ctx, 5)
	})

	assert.NotPanics(t, func() {
		m.ErrorCount.Add(ctx, 2)
	})
}
