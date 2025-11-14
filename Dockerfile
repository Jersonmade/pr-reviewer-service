FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o pr-reviewer-service cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/pr-reviewer-service .

EXPOSE 8080

CMD ["./pr-reviewer-service"]
