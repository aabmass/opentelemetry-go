package foo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestMetricsWork(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
	)
	counter, err := mp.Meter("test").Int64Counter("mycounter")
	require.NoError(t, err)
	counter.Add(ctx, 12)

	rm := &metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, rm))

	t.Logf("Got metrics!: %+v", rm)

	assert.Len(t, rm.ScopeMetrics, 1)
}

func TestOverrideMetricGlobal(t *testing.T) {
	mp1 := sdkmetric.NewMeterProvider()
	mp2 := sdkmetric.NewMeterProvider()

	otel.SetMeterProvider(mp1)
	{
		mp := otel.GetMeterProvider()
		assert.Same(t, mp, mp1)
	}

	otel.SetMeterProvider(mp2)
	{
		mp := otel.GetMeterProvider()
		assert.Same(t, mp, mp2)
	}
}
