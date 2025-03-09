FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go build -o ical-proxy

FROM gcr.io/distroless/base-debian11

WORKDIR /app
COPY --from=builder /app/ical-proxy /app/ical-proxy

EXPOSE 8080
ENTRYPOINT ["/app/ical-proxy"]
