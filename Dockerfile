FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hello/main hello/main.go

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/hello/main .
CMD [ "/app/main" ]
