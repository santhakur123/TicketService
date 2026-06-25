FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ticket-system .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/ticket-system .

EXPOSE 8080

ENV PORT=8080

CMD ["./ticket-system"]
