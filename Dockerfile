FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /sharetrip ./cmd/sharetrip

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /sharetrip .
CMD ["./sharetrip"]
