FROM golang:1.21-bookworm AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/data-insert .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/data-insert /app/data-insert
CMD [ "/app/data-insert" ]
