FROM golang:1.21-bookworm AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/profit-report .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/profit-report /app/profit-report
CMD [ "/app/profit-report" ]
