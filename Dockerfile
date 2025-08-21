FROM golang:1.24.6-alpine3.22 AS build

WORKDIR /app
COPY . .
RUN go build -o /app/kotomi ./cmd/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=build /app/kotomi .

CMD ["./kotomi"]