FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/testra-worker ./cmd/worker

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/testra-worker /usr/local/bin/testra-worker
ENTRYPOINT ["testra-worker"]
