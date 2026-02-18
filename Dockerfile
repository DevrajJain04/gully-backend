# ---- Build stage ----
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gully-backend .

# ---- Run stage ----
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /gully-backend .
COPY .env.example .env
EXPOSE 8080
CMD ["./gully-backend"]
