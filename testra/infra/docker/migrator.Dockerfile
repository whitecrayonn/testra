FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/testra-migrator ./cmd/migrator

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /bin/testra-migrator /usr/local/bin/testra-migrator
COPY --from=builder /app/migrations ./migrations
ENV MIGRATIONS_PATH=migrations
ENTRYPOINT ["testra-migrator"]
