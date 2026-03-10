## การทำ Observability

**Observability** คือ แนวคิดและการปฏิบัติที่ช่วยให้ทีม DevOps, SRE หรือ Developer มองเห็นสิ่งที่เกิดขึ้นภายในระบบได้อย่างชัดเจน ผ่าน **Logs**, **Metrics**, และ **Traces**

- **Metrics** → ค่าตัวเลข เช่น CPU, RAM, จำนวน request, error rate → ช่วยวัดประสิทธิภาพโดยรวม
- **Logs** → ข้อความเหตุการณ์ → ช่วย Debug ปัญหาแบบละเอียด เช่น Stack Trace
- **Traces** → ติดตามเส้นทางการทำงานของ request → ช่วยหาจุด bottleneck ในระบบ distributed

ยิ่งระบบมีหลาย microservices, container หรือรันบน cloud ยิ่งต้องมี Observability ที่ดี เพราะจะช่วย

- ตรวจจับปัญหาได้เร็ว
- แก้ไข root cause ได้ไว
- ลด downtime
- วัด SLA / SLO ได้แม่นยำ

### Observability stack

ปัจจุบันนี้นิยมใช้ **Grafana Stack** (Loki, Tempo, Prometheus) ร่วมกับ **OpenTelemetry Protocol** (OTLP) ซึ่งเป็น protocol มาตรฐาน สำหรับส่ง telemetry data (Traces, Metrics, Logs(บางส่วน))

โฟลว์การทำงานจะเป็นแบบนี้

```bash
[App Container]
   ├─ stdout logs → Promtail → Loki → Grafana
   ├─ OTLP Traces → Otel Collector → Tempo → Grafana
   └─ OTLP Metrics → Otel Collector → Prometheus → Grafana
```

- **OTLP** → ****มาตรฐานส่ง Telemetry (trace, metric)
- **Otel Collector** → รับ OTLP จาก app → ประมวลผล (เช่น batch, filter, enrich) → ส่งต่อไปยัง backend ที่เหมาะสม (Prometheus, Tempo)
- **Promtail** → เก็บ log จากไฟล์/ stdout  ส่งเข้า Loki
- **Loki** → เก็บ **logs** ที่จัดการง่าย คล้าย Prometheus (label-based)
- **Tempo** → เก็บ **traces** เพื่อติดตาม flow ของ request
- **Prometheus** → เก็บ **metrics** แบบ time-series
- **Grafana** → dashboard กลาง เชื่อมทุกตัว ดูได้ในที่เดียว

### ขั้นตอนการทำ Observability

### ปรับรูปแบบของ Zap logger

ในโค้ดตอนนี้ใช้ `ecszap` เพื่อให้ log ที่ออกมามีโครงสร้าง **ECS (Elastic Common Schema)** ซึ่งเหมาะสำหรับส่งไป Elastic Stack

แต่ถ้าจะทำ Observability แบบ OpenTelemetry (**OTel**) จริง ๆ แล้ว **OTel Logging** ยังไม่มี spec ที่เป็น JSON schema มาตรฐานแบบ ECS หรือ OpenTracing สำหรับ trace

ส่วนใหญ่ OTel จะเน้น **logs ต้องมี [semantic attributes](https://opentelemetry.io/docs/specs/semconv/registry/attributes/)** ที่ correlate กับ trace/span context เช่น

- `trace_id` : ID เดียวกันสำหรับ **การเรียก (request)** ทั้งชุด ช่วยให้รู้ว่า log หรือ span ไหนอยู่ใน request เดียวกัน สามารถใช้แทน requestId เดิม หรือใช้คู่กันได้
- `span_id` : ID ของ **จุดการทำงานย่อย (span)** เช่น ใน 1 request อาจมีหลาย span ใช้ span_id แยกว่า log นี้อยู่ในขั้นไหนของ flow
- `service.name` : ชื่อ service หรือ application ใช้แยก log/trace ว่ามาจากระบบไหน
- `severity` : ระดับความรุนแรงของ log เช่น `INFO`, `DEBUG`, `WARN`, `ERROR`
- `timestamp` : เวลาที่เกิด log หรือ span ใช้เรียงลำดับเหตุการณ์ ช่วยวิเคราะห์ timeline ได้ชัดเจน

สิ่งสำคัญคือ **format** ของ log ควรเป็น **JSON + มี field เหล่านี้** เพื่อให้ collector หรือ backend (เช่น Loki) เข้าใจ

เริ่มจากปรับ `zap` ให้รองรับ OpenTelemetry

> แก้ไขไฟล์ `common/logger/logger.go`
>

```go
package logger

import (
 "context"

 "go.opentelemetry.io/otel/trace"
 "go.uber.org/zap"
 "go.uber.org/zap/zapcore"
)

// ...

func Init(serviceName string) (closeLog, error) {
 config := zap.NewProductionConfig() // ใช้ production config → output เป็น JSON

 // ตั้งค่า key ของ field ให้ตรงกับ OTel semantic convention
 config.EncoderConfig.TimeKey = "timestamp"
 config.EncoderConfig.LevelKey = "severity"
 config.EncoderConfig.MessageKey = "message"
 config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

 var err error
 zlog, err := config.Build()

 if err != nil {
  return nil, err
 }

 baseLogger = zlog.With(zap.String("service.name", serviceName))

 return func() error {
  return baseLogger.Sync()
 }, nil
}

// ...

func FromContext(ctx context.Context) *zap.Logger {
 log, ok := ctx.Value(loggerKey{}).(*zap.Logger)
 if !ok {
  log = baseLogger
 }

 // ดึง trace_id + span_id จาก OTel context
 span := trace.SpanContextFromContext(ctx)

 if span.IsValid() {
  return log.With(
   zap.String("trace_id", span.TraceID().String()), // เพิ่ม trace_id เพื่อเชื่อมโยง log กับ trace
   zap.String("span_id", span.SpanID().String()),   // เพิ่ม span_id เพื่อเชื่อมโยง log กับ trace
  )
 }

 return log
}
```

### เก็บ Traces กับ Metrics ด้วย OTLP

สร้าง `TracerProvider` และ `MeterProvider`

> สร้างไฟล์ `common/observability/observability.go`
>

```go
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
```

เพิ่มรับ ENV สำหรับ otel-collector endpoint

> แก้ไขไฟล์ `app/config/config.go`
>

```diff
type Config struct {
 HTTPPort            int
 GracefulTimeout     time.Duration
 DSN                 string
 GatewayHost         string
 GatewayBasePath     string
++ AppName             string
++ OtelExporterEnpoint string
}

func Load() (*Config, error) {
 config := &Config{
  HTTPPort:            env.GetIntDefault("HTTP_PORT", 8090),
  GracefulTimeout:     env.GetDurationDefault("GRACEFUL_TIMEOUT", 5*time.Second),
  DSN:                 env.Get("DB_DSN"),
  GatewayHost:         env.Get("GATEWAY_HOST"),
  GatewayBasePath:     env.GetDefault("GATEWAY_BASEURL", "/api/v1"),
++  AppName:             env.GetDefault("APP_NAME", "go-mma-api"),
++  OtelExporterEnpoint: env.Get("OTEL_EXPORTER_ENDPOINT"),
 }
 err := config.Validate()
 if err != nil {
  return nil, err
 }
 return config, err
}
```

> แก้ไขไฟล์ `app/cmd/main.go`
>

```go
func main() {
 config, err := config.Load()
 if err != nil {
  panic(err.Error())
 }

 closeLog, err := logger.Init(config.AppName)
 if err != nil {
  panic(err.Error())
 }
 defer closeLog()

 shutdown, err := observability.InitOtlp(context.Background(), config.OtelExporterEnpoint, config.AppName)
 if err != nil {
  panic(err.Error())
 }
 defer shutdown(context.Background())

 dbCtx, closeDB, err := sqldb.NewDBContext(config.DSN)
 // ...
}
```

### สร้าง Middleware สำหรับจัดการ Log + Trace + Metric

สร้าง middleware สำหรับสร้าง `log` ใหม่, สร้าง `span` ใหม่ และวัด Metrics ของแต่ละ Request

> สร้างไฟล์ `app/application/middleware/obervability.go`
>

```go
package middleware

import (
 "context"
 "fmt"
 "go-mma/shared/common/logger"
 "runtime/debug"
 "strings"
 "time"

 "github.com/gofiber/fiber/v3"
 "github.com/google/uuid"
 "go.opentelemetry.io/otel"
 "go.opentelemetry.io/otel/attribute"
 "go.opentelemetry.io/otel/codes"
 "go.opentelemetry.io/otel/metric"
 "go.opentelemetry.io/otel/trace"
 "go.uber.org/zap"
)

func Observability() fiber.Handler {

 tracer := otel.GetTracerProvider().Tracer("http_request")
 meter := otel.GetMeterProvider().Meter("http_request")

 // ----- OTel Instruments -----
 requestCounter, _ := meter.Int64Counter("http_requests_total")
 requestDuration, _ := meter.Float64Histogram("http_request_duration_ms")
 inflightCounter, _ := meter.Int64UpDownCounter("http_requests_inflight")
 requestSize, _ := meter.Float64Histogram("http_request_size_bytes")
 responseSize, _ := meter.Float64Histogram("http_response_size_bytes")
 errorCounter, _ := meter.Int64Counter("http_requests_error_total")

 // Skip Paths ที่ไม่ต้องการ trace
 skipPaths := map[string]bool{
  "/health":  true,
  "/metrics": true,
 }
 // กรณีมีการ serve SPA
 staticPrefixes := []string{"/static", "/assets", "/public", "/favicon", "/robots.txt"}

 return func(c fiber.Ctx) error {
  start := time.Now()
  method := c.Method()
  path := c.Path()

  // ตรวจสอบ path ที่เรียกมา
  skip := skipPaths[path]
  for _, prefix := range staticPrefixes {
   if strings.HasPrefix(path, prefix) {
    skip = true
    break
   }
  }

  requestID := c.Get("X-Request-ID")
  if requestID == "" {
   requestID = uuid.New().String()
  }

  // Bind Request ID ลง Response Header
  c.Set("X-Request-ID", requestID)

  var (
   ctx    context.Context
   span   trace.Span
   labels []attribute.KeyValue
  )

  if skip {
   ctx = c.Context()
  } else {
   // สร้าง span ใหม่
   ctx, span = tracer.Start(c.Context(), "HTTP "+method+" "+path,
    // https://opentelemetry.io/docs/specs/semconv/registry/attributes/
    trace.WithAttributes(attribute.String("http.request_id", requestID)),
    trace.WithAttributes(attribute.String("http.request.method", method)),
    trace.WithAttributes(attribute.String("url.path", path)),
   )
   defer span.End()
  }

  // สร้าง child logger
  reqLogger := logger.With(
   zap.String("request_id", requestID),
   zap.String("http.request.method", method),
   zap.String("url.path", path),
  )

  // สร้าง Context ใหม่ที่มี logger
  ctx = logger.NewContext(ctx, reqLogger)
  // แทน Context เดิม
  c.SetContext(ctx)

  // ----- Record Inflight -----
  if !skip {
   inflightCounter.Add(ctx, 1)
  }

  err := c.Next()

  duration := time.Since(start).Milliseconds()
  status := c.Response().StatusCode()

  if !skip {
   labels = []attribute.KeyValue{
    attribute.String("http.request.method", method),
    attribute.String("url.path", path),
    attribute.Int("http.response.status_code", status),
   }

   requestCounter.Add(ctx, 1, metric.WithAttributes(labels...))
   requestDuration.Record(ctx, float64(duration), metric.WithAttributes(labels...))
   inflightCounter.Add(ctx, -1)

   // Request Size (Header Content-Length)
   if reqSize := c.Request().Header.ContentLength(); reqSize > 0 {
    requestSize.Record(ctx, float64(reqSize), metric.WithAttributes(labels...))
   }

   // Response Size (Body Length)
   if resSize := len(c.Response().Body()); resSize > 0 {
    responseSize.Record(ctx, float64(resSize), metric.WithAttributes(labels...))
   }

   if status >= 400 {
    errorCounter.Add(ctx, 1, metric.WithAttributes(labels...))
   }
  }

  // log unhandle error
  if err != nil {
   reqLogger.Error("an error occurred",
    zap.Any("error", err),
    zap.ByteString("stack", debug.Stack()),
   )
  }

  msg := fmt.Sprintf("%d - %s %s", status, method, path)
  reqLogger.Info(msg,
   zap.Int("http.response.status_code", status),
   zap.Int64("duration_ms", duration),
   zap.String("trace_id", span.SpanContext().TraceID().String()), // เพิ่ม trace_id เพื่อเชื่อมโยง log กับ trace
   zap.String("span_id", span.SpanContext().SpanID().String()),   // เพิ่ม span_id เพื่อเชื่อมโยง log กับ trace
  )

  span.SetAttributes(
   attribute.Int("http.response.status_code", status),
  )

  if status >= 400 {
   span.SetStatus(codes.Error, "")
  } else {
   span.SetStatus(codes.Ok, "")
  }

  return err
 }
}

```

เรียกใช้งาน middleware โดยให้ใช้งานเป็นลำดับแรก

> แก้ไฟล์ `app/application/http.go`
>

```diff
func newFiber(config config.Config) *fiber.App {
 app := fiber.New(fiber.Config{
  AppName: fmt.Sprintf("Go MMA version %s", build.Version),
 })

 // global middleware
++ app.Use(middleware.Observability()) // จัดการ log + trace + metric
 app.Use(cors.New())                               // CORS ลำดับแรก เพื่อให้ OPTIONS request ผ่านได้เสมอ
-- app.Use(requestid.New())                          // ไม่ต้องใช้แล้ว
 app.Use(recover.New())                            // auto-recovers from panic (internal only)
-- app.Use(middleware.RequestLogger())               // ไม่ต้องใช้แล้ว
 app.Use(middleware.ResponseError())

 app.Get("/docs/*", middleware.APIDoc(config))

 app.Get("/", func(c fiber.Ctx) error {
  return c.JSON(map[string]string{"version": build.Version, "time": build.Time})
 })

 return app
}
```

### ติดตามเส้นทางการทำงานของ request ด้วย Traces

การติดตามว่าทำงานผ่านฟังก์ชันไหนบ้าง ทำได้โดยการสร้าง span ใหม่ขึ้นมา ตัวอย่าง

```go
tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("http_handler")
ctx, span := tracer.Start(ctx, "Endpoint:CreateCustomer")
defer span.End()
```

ตัวอย่างการ trace ในการสร้าง customer

- `customer/internal/feature/create/endpoint.go`

    ```diff
    func createCustomerHTTPHandler(c fiber.Ctx) error {
    ++ ctx := c.Context()
    ++ tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("http_handler")
    ++ ctx, span := tracer.Start(ctx, "Endpoint:CreateCustomer")
    ++ defer span.End()
    
     // แปลง request body -> dto
     var req CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      // จัดการ error response ที่ middleware
      return errs.InputValidationError(err.Error())
     }
     
    -- logger.FromContext(c.Context()).Info(fmt.Sprintf("Received customer: %v", req))
    ++  // ต้องใช้ ctx ของ span เพื่อใส่เลข span ใหม่ลงใน logger
    ++ logger.FromContext(ctx).Info(fmt.Sprintf("Received customer: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
      // จัดการ error response ที่ middleware
      return errs.InputValidationError(err.Error())
     }
    
     // *** ส่งไปที่ Command Handler แทน Service ***
     resp, err := mediator.Send[*CreateCustomerCommand, *CreateCustomerCommandResult](
    --  c.Context(),
    ++  ctx, // ต้องส่ง ctx ของ span ไป เพื่อสร้าง span ต่อ
      &CreateCustomerCommand{CreateCustomerRequest: req},
     )
    
     // จัดการ error จาก Service Layer หากเกิดขึ้น
     if err != nil {
      // จัดการ error response ที่ middleware
      return err
     }
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    ```

- `customer/internal/feature/create/command_handler.go`

    ```go
    func (h *createCustomerCommandHandler) Handle(ctx context.Context, cmd *CreateCustomerCommand) (*CreateCustomerCommandResult, error) {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("command_handler")
     ctx, span := tracer.Start(ctx, "Handle:CreateCustomerCommand")
     defer span.End()
     
     // ...
    }
    ```

- `customer/internal/repository/customer.go`

    ```go
    func (r *customerRepository) Create(ctx context.Context, customer *model.Customer) error {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("repository")
     ctx, span := tracer.Start(ctx, "Repository:CustomerRepository:Create")
     defer span.End()
    
     // ...
    }
    
    func (r *customerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("repository")
     ctx, span := tracer.Start(ctx, "Repository:CustomerRepository:ExistsByEmail")
     defer span.End()
    
     // ...
    }
    
    func (r *customerRepository) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("repository")
     ctx, span := tracer.Start(ctx, "Repository:CustomerRepository:FindByID")
     defer span.End()
    
     // ...
    }
    
    func (r *customerRepository) FindByIDForUpdate(ctx context.Context, id int64) (*model.Customer, error) {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("repository")
     ctx, span := tracer.Start(ctx, "Repository:CustomerRepository:FindByIDForUpdate")
     defer span.End()
    
     // ...
    }
    
    func (r *customerRepository) UpdateCredit(ctx context.Context, m *model.Customer) error {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("repository")
     ctx, span := tracer.Start(ctx, "Repository:CustomerRepository:UpdateCredit")
     defer span.End()
    
     // ...
    }
    ```

- `customer/internal/domain/eventhandler/customer_created_handler.go`

    ```go
    func (h *customerCreatedDomainEventHandler) Handle(ctx context.Context, evt domain.DomainEvent) error {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("domain_event")
     ctx, span := tracer.Start(ctx, "DomainEvent:CreateCustomer")
     defer span.End()
     
     // ...
    }
    ```

- `notification/internal/integration/customer/welcome_email_handler.go`

    ```go
    func (h *welcomeEmailHandler) Handle(ctx context.Context, evt eventbus.Event) error {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("integration_event")
     ctx, span := tracer.Start(ctx, "IntegrationEvent:WelcomeEmail")
     defer span.End()
     
     // ...
    }
    ```

- `notification/service/notification.go`

    ```go
    func (s *notificationService) SendEmail(ctx context.Context, to string, subject string, payload map[string]any) error {
     tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("service")
     ctx, span := tracer.Start(ctx, "Service:SendEmail")
     defer span.End()
     
     // ...
    }
    ```

### สร้างไฟล์ config

สร้างไฟล์ config สำหรับการเก็บ Log, Trace และ Metric

- สร้างไฟล์ `config/promtail-config.yml` สำหรับดึง log ส่งเข้า Loki

    ```yaml
    # ตัวอย่าง promtail-config.yml
    server:
      http_listen_port: 9080
      grpc_listen_port: 0
    
    positions:
      filename: /tmp/positions.yaml
    
    clients:
      - url: http://loki:3100/loki/api/v1/push
    
    scrape_configs:
      - job_name: docker-logs
        docker_sd_configs:
          - host: unix:///var/run/docker.sock
            refresh_interval: 5s
            filters:
              - name: label
                values: ["logging=promtail"]
    
        relabel_configs:
          - source_labels: ['__meta_docker_container_name']
            regex: '/(.*)'
            target_label: 'container'
          - source_labels: ['__meta_docker_container_log_stream']
            target_label: 'logstream'
          - source_labels: ['__meta_docker_container_label_logging_jobname']
            target_label: 'job'
    
        pipeline_stages:
          - docker: {}
          - json:
              expressions:
                service.name:
                severity:
                message:
                request_id:
                trace_id:
                span_id:
          - labels:
              app_name:
              severity:
              request_id:
              trace_id:
              span_id:
    ```

- สร้างไฟล์ `config/tempo.yml`

    ```yaml
    auth_enabled: false
    
    stream_over_http_enabled: true
    server:
      http_listen_port: 3200
      log_level: info
    
    distributor:
      receivers:
        otlp:
          protocols:
            grpc:
              endpoint: "tempo:4317"
    
    ingester:
      trace_idle_period: 10s
      max_block_duration: 5m
    
    compactor:
      compaction:
        block_retention: 1h
    
    storage:
      trace:
        backend: local
        local:
          path: /tmp/tempo/blocks
        wal:
          path: /tmp/tempo/wal
    
    ```

- สร้างไฟล์ `config/prometheus.yml`

    ```yaml
    global:
      scrape_interval: 5s
    
    scrape_configs:
      - job_name: 'otel-collector'
        static_configs:
          - targets: ['otel-collector:9464']
    ```

- สร้างไฟล์ `config/otel-collector-config.yml` สำหรับให้ exporter ส่ง trace เข้า tempo และ metric เข้า prometheus

    ```yaml
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318
    
    exporters:
      otlp/tempo:
        endpoint: tempo:4317
        tls:
          insecure: true
      # Expose Prometheus endpoint → Prometheus server จะ scrape ตรงนี้
      prometheus:
        endpoint: "0.0.0.0:9464"
      # ดู stdout/stderr ของ Collector (log console) เพื่อ debug
      debug:
        # normal → ข้อมูลหลัก (resource, scope, metric name)
        # detailed → ลึกสุด! เห็น datapoints ทุกรายการ
        verbosity: detailed
    
    processors:
      batch:
    
    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [batch]
          exporters: [otlp/tempo]
          # อย่าเปิดใน production → log วิ่งมหาศาล
          # exporters: [otlp/tempo, debug]
        metrics:
          receivers: [otlp]
          processors: [batch]
          exporters: [prometheus]
          # อย่าเปิดใน production → log วิ่งมหาศาล
          # exporters: [prometheus, debug]
    ```

### เพิ่ม Grafana Stack ใน Docker Compose

แก้ไขไฟล์ `docker-compose.prod.yml`

```yaml
networks:
  # network สำหรับ reverse proxy
  frontend:
  # network สำหรับ services ภายใน
  backend:

volumes:
  loki_data:
  grafana-data:
  prometheus-data:

services:
  loki:
    image: grafana/loki:latest
    container_name: loki
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - loki_data:/loki
    # ports:
    #   - "3100:3100" 
    networks:
      - backend

  promtail:
    image: grafana/promtail:latest
    container_name: promtail
    volumes:
      - ./config/promtail-config.yml:/etc/promtail/promtail-config.yml
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    command: -config.file=/etc/promtail/promtail-config.yml
    networks:
      - backend
    depends_on:
      - loki

  otel-collector:
    image: otel/opentelemetry-collector:latest
    container_name: otel-collector
    volumes:
      - ./config/otel-collector-config.yaml:/etc/otelcol/config.yaml
    command: ["--config=/etc/otelcol/config.yaml"]
    ports:
      - "4317:4317"   # OTLP gRPC
    #   - "4318:4318"   # OTLP HTTP
    #   - "9464:9464"   # Prometheus scrape metrics endpoint
    #   - "13133:13133" # Health check (extension)
    networks:
      - backend
    depends_on:
      - tempo

  tempo:
    image: grafana/tempo:latest
    container_name: tempo
    volumes:
      - ./config/tempo.yaml:/etc/tempo.yaml
    command: [ "-config.file=/etc/tempo.yaml" ]
    # ports:
    #   - "3200" # tempo
    #   - "4317" # otlp grpc
    networks:
      - backend

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    command:
      - --config.file=/etc/prometheus/prometheus.yml
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    # ports:
    #   - "9090:9090"
    networks:
      - backend

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    networks:
      - backend
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
    depends_on:
      - loki
      - tempo
      - prometheus

  db:
    environment:
      POSTGRES_DB: go-mma-db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    networks:
      - backend
    # ตรวจสอบว่าพร้อมทำงานรึยัง
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 1s
      timeout: 5s
      retries: 10

  migrate:
    image: migrate/migrate:v4.18.3
    volumes:
      - ./migrations:/migrations
    networks:
      - backend
    depends_on:
      # ต้องรันหลังจากที่ db พร้อมทำงานแล้ว
      db:
        condition: service_healthy
    # สั่ง migrate up
    command: -verbose -path=/migrations/ -database "postgres://postgres:postgres@db:5432/go-mma-db?sslmode=disable" up

  app:
    build: .
    image: go-mma-api
    container_name: go-mma-api
    environment:
      - GRACEFUL_TIMEOUT=5s
      - DB_DSN=postgres://postgres:postgres@db:5432/go-mma-db?sslmode=disable
      - GATEWAY_HOST=localhost
      - GATEWAY_BASEURL=/api/v1
      - OTEL_EXPORTER_ENDPOINT=otel-collector:4317
    networks:
      - frontend
      - backend
    depends_on:
      db:
        # ต้องรันหลังจากที่ db พร้อมทำงานแล้ว
        condition: service_healthy 
      migrate:
        # ต้องรันหลังจากที่ migrate ทำงานสำเร็จแล้ว
        condition: service_completed_successfully
      otel-collector:
        # ต้องรันหลังจากที่ otel-collector เริ่มทำงานแล้ว
        condition: service_started
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "5"
    # เพิ่ม label สำหรับ filtering logs
    labels:
      logging: "promtail"
      logging_jobname: "containerlogs"

  proxy:
    image: nginx:1.29.0-alpine
    container_name: nginx-proxy
    ports:
      - "80:80"
    volumes:
      # เรียกใช้ config/nginx.conf
      - ./config/nginx.conf:/etc/nginx/nginx.conf:ro
    networks:
      - frontend
    depends_on:
      app:
        # ต้องรันหลังจากที่ app เริ่มทำงานแล้ว
        condition: service_started
```

### ทดสอบ

- Build image ใหม่

    ```bash
    BUILD_VERSION=latest make image
    ```

- รัน docker compose

    ```bash
    make produp
    ```

- เข้า Grafana ที่ <http://localhost:3000> ใช้ username: admin และ password: admin
- เพิ่ม Data sources
  - Loki → <http://loki:3100>
  - Tempo → <http://tempo:3200> และเปิดใช้งาน Trace to logs เลือกไปที่ loki และเปิด Filter by trace ID กับ Filter by span ID
  - Prometheus → <http://prometheus:9090>
- ทดสอบเรียก API จากไฟล์ `customers.http`
- สามารถดู Log, Trace และ Metric ได้จากเมนู Explore <http://localhost:3000/explore>

---

