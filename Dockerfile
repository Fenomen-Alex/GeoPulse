# Stage 1: Build SolidJS Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/client
COPY client/package*.json ./
RUN npm ci
COPY client/ ./
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:1.26-alpine AS backend-builder
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
# Copy compiled frontend into Go embedding directory
COPY --from=frontend-builder /app/client/dist ./cmd/server/public
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /geopulse ./cmd/server/main.go

# Stage 3: Final Minimal Runtime Image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /geopulse /app/geopulse

EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/app/geopulse"]
