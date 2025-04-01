package main

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"net/http"
	"pay/middleware"
	"pay/pkg/tracing"
	"time"
)

func main() {
	// Инициализируем OpenTelemetry (tracer provider) для pay
	closeFn := tracing.InitTracerProvider("pay")
	defer closeFn()

	// Устанавливаем единый propagator (W3C TraceContext)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	r := gin.Default()

	// Если контекст еще не был извлечен, можно добавить middleware, который делает Extract:
	r.Use(func(c *gin.Context) {
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Используем middleware CreateSpanByTraceID для создания дочернего спана
	// Этот middleware проверит наличие валидного родительского спана в контексте и, если он есть, создаст дочерний спан.
	r.Use(middleware.CreateSpanByParent("pay-handle-/pay/test", "pay-tracer"))

	// Обработчик, в котором не нужно явно создавать спан – он уже создан middleware
	r.GET("/pay/test", func(c *gin.Context) {
		log.Println("pay service: received request /pay/test")
		time.Sleep(2 * time.Second)
		c.String(http.StatusOK, "pay service OK")
	})

	log.Println("pay service is running on :9090")
	if err := r.Run(":9090"); err != nil {
		log.Fatalf("Failed to start pay service: %v", err)
	}

	//// Инициализируем OpenTelemetry (tracer provider) для pay
	//closeFn := tracing.InitTracerProvider("pay")
	//defer closeFn()
	//
	//r := gin.Default()
	//
	//r.GET("/pay/test", func(c *gin.Context) {
	//	// Извлекаем trace-контекст из заголовков, переданных от router
	//	ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
	//	tracer := otel.Tracer("pay-tracer")
	//
	//	// Создаем server span, используя извлеченный context
	//	ctx, span := tracer.Start(ctx, "pay-handle-/pay/test")
	//	defer span.End()
	//
	//	log.Println("pay service: received request /pay/test")
	//	time.Sleep(2 * time.Second)
	//
	//	c.String(http.StatusOK, "pay service OK")
	//})
	//
	//log.Println("pay service is running on :9090")
	//if err := r.Run(":9090"); err != nil {
	//	log.Fatalf("Failed to start pay service: %v", err)
	//}
}
