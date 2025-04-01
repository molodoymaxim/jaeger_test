package main

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"router/middleware"
	"router/pkg/tracing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	pb "github.com/molodoymaxim/jaeger_test/pay_grpc/pkg/pay_grpc/v1"
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

	r.GET("/testgrpc", func(c *gin.Context) {
		ctx := c.Request.Context()
		err := callPayGRPC(ctx)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to call pay_grpc: %v", err)
			return
		}
		c.String(http.StatusOK, "test grpc end")
	})

	log.Println("router is running on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start router: %v", err)
	}

	log.Println("router is running on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start router: %v", err)
	}
}

func callPayGRPC(ctx context.Context) error {
	conn, err := grpc.Dial("pay-grpc:9091", grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewPayServiceClient(conn)

	// Инжектируем trace-контекст в grpc.Metadata.
	md := metadata.New(nil)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(md))
	outCtx := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.Test(outCtx, &pb.TestRequest{})
	if err != nil {
		return err
	}
	log.Printf("pay_grpc response: %s", resp.Message)
	return nil
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
