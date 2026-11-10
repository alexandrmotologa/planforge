# Stage 1: Build React 19 Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o planforge ./cmd/planforge

# Stage 3: Minimal Scratch Runtime Image
FROM alpine:3.21
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=backend-builder /app/planforge /usr/local/bin/planforge
EXPOSE 8080
ENTRYPOINT ["planforge"]
CMD ["serve", "--host", "0.0.0.0", "--port", "8080"]
