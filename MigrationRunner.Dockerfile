FROM golang:1.20 as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal

RUN go test -v ./cmd/... ./internal/...

RUN go build -mod=readonly -v -o runner ./cmd/migration_runner/runner.go

FROM gcr.io/distroless/base as deploy

WORKDIR /app
COPY --from=build app/runner .

# Must be specified in vector format, due to distroless not having a shell
# https://github.com/GoogleContainerTools/distroless#entrypoints
ENTRYPOINT ["./runner"]