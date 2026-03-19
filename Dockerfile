FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o modbus-driver ./cmd/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/modbus-driver .
COPY ./cmd/.env .
CMD ["./modbus-driver"]