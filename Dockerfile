# https://docs.docker.com/engine/reference/builder/
FROM golang:1.15-alpine AS builder
WORKDIR /app
ENV CGO_ENABLED=0
COPY . /app
RUN go build -o /signaller /app/cmd/signaller/...

FROM scratch
CMD ["/signaller"]
COPY --from=builder /signaller /signaller
