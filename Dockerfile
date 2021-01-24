# https://docs.docker.com/engine/reference/builder/
FROM golang AS builder
WORKDIR /app
COPY . .
RUN go build -o /tmp/signaller ./cmd/signaller/...

FROM scratch
COPY --from=builder /tmp/signaller /bin/signaller
ENTRYPOINT ["signaller"]