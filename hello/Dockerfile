FROM golang:1.21-bookworm AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main /app/hello
CMD [ "/app/hello" ]
