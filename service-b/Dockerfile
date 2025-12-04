FROM golang:1.23.5 AS builder
WORKDIR /app
COPY service-b/go.mod service-b/go.sum ./
COPY pkg/ ../pkg/
RUN go mod download
COPY service-b/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o service-b cmd/api/main.go

FROM alpine:3.22
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/service-b .
CMD ["./service-b"]
