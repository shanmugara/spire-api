# syntax=docker/dockerfile:experimental
FROM golang:1.23 AS build

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV SPIFFE_ENDPOINT_SOCKET=unix:///run/spire/sockets/api.sock

WORKDIR /work
COPY . /work

RUN --mount=type=cache,target=/root/.cache/go-build,sharing=private \
  go build -o bin/spire-api .

# ---
FROM scratch AS run

ENV SPIFFE_ENDPOINT_SOCKET=unix:///run/spire/sockets/api.sock

COPY --from=build /work/bin/spire-api /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/spire-api"]