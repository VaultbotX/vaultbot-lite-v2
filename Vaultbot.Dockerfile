FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal

RUN go build -mod=readonly -v -o vaultbot ./cmd/vaultbot/vaultbot.go

FROM gcr.io/distroless/base AS deploy

WORKDIR /app
COPY --from=build app/vaultbot .

# Must be specified in vector format, due to distroless not having a shell
# https://github.com/GoogleContainerTools/distroless#entrypoints
ENTRYPOINT ["./vaultbot"]