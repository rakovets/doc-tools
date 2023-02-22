FROM golang:1.20-alpine AS builder
WORKDIR /build
RUN go version
COPY go.mod ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o main ./cmd/app/main.go

FROM pandoc/core:latest
WORKDIR /app
ARG CONFIG_PATH=configs/config.yaml
ENV CONFIG_PATH=$CONFIG_PATH
COPY --from=builder /build/main ./
ENTRYPOINT ["/app/main"]
