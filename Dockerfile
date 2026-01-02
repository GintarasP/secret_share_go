FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build -o server ./cmd/server

FROM alpine:latest  

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/static ./static
COPY --from=builder /app/api ./api

EXPOSE 8080

CMD ["./server"]
