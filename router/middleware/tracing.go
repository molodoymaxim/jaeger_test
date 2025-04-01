package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"log"
)

func CreateTracerAndSpan(operationName, traceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := otel.Tracer(traceName)

		// Start попробует извлечь родительский спанКонтекст и если родителя не будет, то создаст новый
		ctx, span := tracer.Start(c.Request.Context(), operationName)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func CreateSpanByParent(spanName, tracerName string) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Получаем родительский спан из текущего контекста.
		parentSpan := trace.SpanFromContext(c.Request.Context())
		if !parentSpan.SpanContext().IsValid() {
			log.Println("Не найден валидный родительский спан в контексте")
			c.Next()
			return
		}

		// Используем стандартный метод для создания дочернего спана.
		tracer := otel.Tracer(tracerName)
		ctx, span := tracer.Start(c.Request.Context(), spanName)
		defer span.End()

		// Обновляем контекст запроса.
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
