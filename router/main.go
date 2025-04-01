package main

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"router/middleware"
	"router/pkg/tracing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	// Инициализируем OpenTelemetry (tracer provider)
	closeFuncTP := tracing.InitTracerProvider("router")
	defer closeFuncTP()

	r := gin.Default()
	r.Use(middleware.CreateTracerAndSpan("router-handle-/test", "router-tracer"))
	//r.Use(middleware.CreateSpanByTraceID("child-span-from-tracer_id", "router-tracer"))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "router ok")
	})

	r.GET("/test", func(c *gin.Context) {
		// Контекст запроса уже содержит спаны, созданные в middleware
		ctx := c.Request.Context()

		// Пример: имитируем время обработки
		time.Sleep(500 * time.Millisecond)

		// Вызываем функцию, которая сделает запрос к сервису pay,
		// прокидывая текущий трейс-контекст.
		err := queryToParserWithContextService(
			ctx,
			"GET",
			"/pay/test", // endpoint сервиса pay
			map[string]string{
				"cache-control": "no-cache",
				"Content-Type":  "application/json",
			},
			nil,
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to call pay: %v", err)
			return
		}

		c.String(http.StatusOK, "test end")
	})

	log.Println("router is running on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start router: %v", err)
	}

	// Для примера запускаем HTTP-сервер на порту 8080
	//mux := http.NewServeMux()
	//mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
	//	w.Write([]byte("router pong"))
	//})

	// Эндпоинт /test — простая ручка, из которой мы вызываем queryToParserWithContextService
	//testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	ctx := r.Context()
	//
	//	// Пример: время обработки
	//	time.Sleep(500 * time.Millisecond)
	//
	//	// Вызываем нашу функцию, которая сделает запрос к pay
	//	err := queryToParserWithContextService(
	//		ctx,
	//		"GET",
	//		"/pay/test", // endpoint у pay
	//		map[string]string{
	//			"cache-control": "no-cache",
	//			"Content-Type":  "application/json",
	//		},
	//		nil,
	//	)
	//	if err != nil {
	//		http.Error(w, "failed to call pay: "+err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//
	//	w.Write([]byte("test end"))
	//})

	//log.Println("router is running on :8081")
	//if err := http.ListenAndServe(":8081", mux); err != nil {
	//	log.Fatalf("Failed to start router: %v", err)
	//}
}

// queryToParserWithContextService делает HTTP-запрос к сервису pay
// (или другому сервису) с прокидыванием текущего контекста трейсинга.
func queryToParserWithContextService(ctx context.Context, method string, endpoint string, headers map[string]string, body []byte) error {
	// Создаём вложенный span
	//tracer := otel.Tracer("router-tracer")
	//ctx, span := tracer.Start(ctx, "queryToParserWithContextService")
	//defer span.End()
	//middleware.CreateSpanByTraceID("child-span-from-tracer_id", "router-tracer")

	// Формируем запрос
	url := "http://pay:9090" + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		//span.RecordError(err)
		return err
	}

	// Добавляем хедеры, если нужно
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Важно: внедрить (Inject) контекст трейсинга в заголовки запроса
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		//span.RecordError(err)
		return err
	}
	defer resp.Body.Close()

	log.Printf("Pay response status: %s", resp.Status)
	return nil
}

// initTracerProvider настраивает экспорт трейсов в Jaeger
//func initTracerProvider() *sdktrace.TracerProvider {
//	endpoint := os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT")
//	if endpoint == "" {
//		endpoint = "http://localhost:14268/api/traces"
//	}
//
//	exp, err := jaeger.New(
//		jaeger.WithCollectorEndpoint(
//			jaeger.WithEndpoint(endpoint),
//		),
//	)
//	if err != nil {
//		log.Fatalf("failed to create jaeger exporter: %v", err)
//	}
//
//	serviceName := os.Getenv("OTEL_SERVICE_NAME")
//	if serviceName == "" {
//		serviceName = "router"
//	}
//
//	r, err := resource.New(
//		context.Background(),
//		resource.WithAttributes(
//			semconv.ServiceNameKey.String(serviceName),
//		),
//	)
//	if err != nil {
//		log.Fatalf("failed to create resource: %v", err)
//	}
//
//	tp := sdktrace.NewTracerProvider(
//		sdktrace.WithBatcher(exp),
//		sdktrace.WithResource(r),
//	)
//	// Регистрируем провайдер как глобальный
//	otel.SetTracerProvider(tp)
//
//	// Настраиваем пропагацию (прокидку) контекста трейсинга через HTTP-заголовки
//	otel.SetTextMapPropagator(propagation.TraceContext{})
//
//	return tp
//}
