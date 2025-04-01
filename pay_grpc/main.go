package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	// Импортируем пакет для инициализации TracerProvider для pay_grpc.
	"github.com/molodoymaxim/jaeger_test/pay_grpc/pkg/tracing"

	// Импортируем сгенерированные proto-файлы.
	pb "github.com/molodoymaxim/jaeger_test/pay_grpc/pkg/pay_grpc/v1"
)

// server реализует интерфейс pb.PayServiceServer.
type server struct {
	pb.UnimplementedPayServiceServer
}

// Test – обработчик gRPC-запроса.
// Он извлекает trace-контекст из входящего gRPC-контекста (metadata),
// создает server span и имитирует обработку запроса.
func (s *server) Test(ctx context.Context, req *pb.TestRequest) (*pb.TestResponse, error) {
	// Извлекаем metadata из входящего контекста.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	// Преобразуем metadata в carrier для propagator.
	carrier := metadataCarrier(md)
	// Извлекаем trace-контекст из carrier.
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Создаем новый server span, который наследует trace_id, извлеченный из metadata.
	tracer := otel.Tracer("pay-grpc-tracer")
	ctx, span := tracer.Start(ctx, "pay-grpc-handle-Test")
	defer span.End()

	log.Println("pay_grpc service: received Test request")
	// Имитация обработки.
	time.Sleep(2 * time.Second)

	return &pb.TestResponse{
		Message: "pay_grpc OK",
	}, nil
}

// metadataCarrier адаптирует gRPC metadata к интерфейсу propagation.TextMapCarrier.
type metadataCarrier metadata.MD

func (mc metadataCarrier) Get(key string) string {
	values := metadata.MD(mc).Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (mc metadataCarrier) Set(key, value string) {
	metadata.MD(mc)[key] = []string{value}
}

func (mc metadataCarrier) Keys() []string {
	out := make([]string, 0, len(mc))
	for k := range mc {
		out = append(out, k)
	}
	return out
}

func main() {
	// Инициализируем OpenTelemetry для pay_grpc.
	shutdown := tracing.InitTracerProvider("pay_grpc")
	defer shutdown()

	// Устанавливаем propagator для использования W3C TraceContext.
	otel.SetTextMapPropagator(propagation.TraceContext{})

	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPayServiceServer(grpcServer, &server{})

	log.Println("pay_grpc service is running on :9091")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
