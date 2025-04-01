FROM golang:1.22.1-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o pay main.go

CMD ["./pay"]
