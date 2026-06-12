FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags "-s -w" -o backend-app cmd/api/main.go

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Jakarta

WORKDIR /app

COPY --from=builder /app/backend-app .

EXPOSE 3002

CMD ["./backend-app"]