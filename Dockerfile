FROM golang:tip-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root

COPY --from=builder /app/main .

COPY --from=builder /app/static ./static

ENV STAGE=prod
ENV APP_ENV=prod

EXPOSE 8000

CMD ["./main"]