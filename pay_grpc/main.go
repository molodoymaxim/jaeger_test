package main

import (
	"context"
	"log"
	"net"
	"time"

	//pb "example.com/paygrpc/proto" // Замените на путь к вашим сгенерированным файлам proto

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"

	"pay_grpc/pkg/tracing" // Общий пакет для инициализации tracer provider
)

// server реализует pb.PayServiceServer.
type server struct {
	pb.UnimplementedPayServiceServer
}

// Test – простой метод, который имитирует обработку запроса.
func (s *server) Test(ctx context.Context, req *pb.TestRequest) (*pb.TestResponse, error) {
	// Извлекаем trace-контекст из grpc metadata.
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrierFromGRPC(ctx))
	tracer := otel.Tracer("pay-grpc-tracer")
	ctx, span := tracer.Start(ctx, "pay-grpc-handle-Test")
	defer span.End()

	log.Println("pay-grpc: received Test request")
	time.Sleep(2 * time.Second)
	return &pb.TestResponse{Message: "pay-grpc OK"}, nil
}

func main() {
	// Инициализируем OpenTelemetry для pay-grpc.
	shutdown := tracing.InitTracerProvider("pay-grpc")
	defer shutdown()

	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPayServiceServer(grpcServer, &server{})

	log.Println("pay-grpc service is running on :9091")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
