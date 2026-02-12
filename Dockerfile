FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o geoswitch ./cmd/geoswitch

FROM alpine:latest
WORKDIR /
COPY --from=builder /app/geoswitch /usr/local/bin/geoswitch

EXPOSE 8080
ENTRYPOINT ["geoswitch"]
