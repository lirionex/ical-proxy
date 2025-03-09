FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o ./ical-proxy

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/ical-proxy /app/ical-proxy

EXPOSE 8080
ENTRYPOINT ["/app/ical-proxy"]
