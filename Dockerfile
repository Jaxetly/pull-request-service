FROM golang:1.25-alpine as builder
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /main main.go

FROM alpine:3
WORKDIR /app 
COPY --from=builder main /app/main
COPY openapi.yml /app/openapi.yml 
ENTRYPOINT ["/app/main"]