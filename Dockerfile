# https://docs.docker.com/engine/reference/builder/
FROM golang AS builder
WORKDIR /app
COPY . /app
RUN ls -lah; go build -o /signaller /app/cmd/signaller/...

FROM scratch
ENTRYPOINT ["/signaller"]
COPY --from=builder /signaller /signaller
