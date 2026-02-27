FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ramen ./cmd/ramen

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /bin/ramen /usr/local/bin/

ENTRYPOINT ["ramen"]

CMD ["-h"]
