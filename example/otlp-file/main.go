package main

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var _ otlpmetric.Client = (*fileClient)(nil)

type fileClient struct {
	fileName string
}

func (f *fileClient) UploadMetrics(ctx context.Context, rms *mpb.ResourceMetrics) error {
	file, err := os.OpenFile(f.fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := protojson.Marshal(rms)
	if err != nil {
		return err
	}
	bytes = append(bytes, byte('\n'))

	if _, err = file.Write(bytes); err != nil {
		return err
	}
	return nil
}

func (f *fileClient) ForceFlush(_ context.Context) error {
	return nil
}
func (f *fileClient) Shutdown(_ context.Context) error {
	return nil
}

func main() {
	ctx := context.Background()
	exp := otlpmetric.New(&fileClient{fileName: "./otlp.json"})
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(time.Second))))
	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	global.SetMeterProvider(meterProvider)

	counter, err := meterProvider.Meter("foo").SyncInt64().Counter("Foo")
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(time.Millisecond * 250)
	go func() {
		for {
			<-ticker.C
			counter.Add(ctx, 12, attribute.String("attr1", "v1"))
			counter.Add(ctx, 15, attribute.String("attr1", "v2"))
		}
	}()

	timer := time.NewTimer(time.Second * 3)
	<-timer.C
}
