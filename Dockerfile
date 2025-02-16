FROM golang:1.22 as builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o avito-app ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/avito-app .
COPY --from=builder /app/migrations ./migrations

CMD ["./avito-app"]