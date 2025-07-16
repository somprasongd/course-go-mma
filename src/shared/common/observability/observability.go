package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type shutdown func(ctx context.Context) error

func InitOtlp(ctx context.Context, collectorAddr, serviceName string) (shutdown, error) {
	// ถ้าไม่ระบุ collector address ให้ไม่ต้องใช้งาน trace กับ metric
	if len(collectorAddr) == 0 {
		return func(ctx context.Context) error { return nil }, nil
	}

	// Resource: บอกข้อมูล service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// OTLP trace exporter → otel-collector endpoint
	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(collectorAddr),
		otlptracegrpc.WithInsecure(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// OTLP metric exporter → otel-collector endpoint
	metricExp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(collectorAddr),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExp),
		),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// Start runtime instrumentation → เช่น memory, GC, goroutines
	runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(10 * time.Second),
	)

	return func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown tracer: %w", err)
		}

		if err := mp.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown meter: %w", err)
		}

		return nil
	}, nil
}
