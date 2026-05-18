# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /mhentai-backend .

# Run stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /mhentai-backend /app/mhentai-backend

EXPOSE 8080
ENV PORT=8080
CMD ["/app/mhentai-backend"]
