FROM golang:1.24.6-alpine3.22 AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/kotomi ./cmd/main.go

FROM alpine:latest
WORKDIR /app

# Create directory for database storage
RUN mkdir -p /app/data

COPY --from=build /app/kotomi .

# Declare volume for database persistence
VOLUME ["/app/data"]

# Set default DB_PATH to use the volume
ENV DB_PATH=/app/data/kotomi.db

CMD ["./kotomi"]