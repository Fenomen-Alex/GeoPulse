# Stage 1: Build SolidJS Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/client

ARG VITE_GOOGLE_CLIENT_ID
ARG VITE_TEST_MODE=true

ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID
ENV VITE_TEST_MODE=$VITE_TEST_MODE

COPY client/package*.json ./
RUN npm ci
COPY client/ ./
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:1.26-alpine AS backend-builder
WORKDIR /app/server

# Install musl build tools required for go-libsql CGO compilation on Alpine
RUN apk add --no-cache gcc musl-dev

COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
COPY --from=frontend-builder /app/client/dist ./cmd/server/public

# Build natively for host architecture
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -trimpath -o /geopulse ./cmd/server/main.go

# Stage 3: Final Minimal Runtime Image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /geopulse /app/geopulse

EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/app/geopulse"]