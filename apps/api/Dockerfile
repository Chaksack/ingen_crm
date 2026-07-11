# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN go build -o /bin/api

FROM alpine:3.19
WORKDIR /app
ENV API_PORT=8080
COPY --from=build /bin/api /app/api
EXPOSE 8080
CMD ["/app/api"]
