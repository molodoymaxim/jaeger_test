package middleware

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"log"
)

// CreateTracerAndSpan создает (или продолжает) трейс и добавляет базовый спан в gRPC‑контекст.
// Используйте этот interceptor для серверных вызовов, чтобы новый спан оборачивал вызов метода.
func CreateTracerAndSpan(operationName, tracerName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tracer := otel.Tracer(tracerName)
		newCtx, span := tracer.Start(ctx, operationName)
		// span.End() вызовется после завершения обработки запроса
		defer span.End()

		// Вызываем следующий обработчик с обновленным контекстом
		return handler(newCtx, req)
	}
}

// CreateSpanByParent создает дочерний спан, используя родительский спан из gRPC‑контекста.
// Если родительский спан не найден, просто продолжает обработку без создания нового спана.
func CreateSpanByParent(spanName, tracerName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Извлекаем родительский спан из входящего контекста.
		parentSpan := trace.SpanFromContext(ctx)
		if !parentSpan.SpanContext().IsValid() {
			log.Println("Не найден валидный родительский спан в контексте")
			return handler(ctx, req)
		}

		tracer := otel.Tracer(tracerName)
		newCtx, span := tracer.Start(ctx, spanName)
		defer span.End()

		return handler(newCtx, req)
	}
}
