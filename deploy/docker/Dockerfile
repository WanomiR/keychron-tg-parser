FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod init github.com/wanomir/keychron-tg-parser \
    && go mod tidy

RUN go build -o ./main ./cmd/app

FROM alpine:latest

COPY --from=builder /app/main /main

EXPOSE 8888

CMD ["/main"]
