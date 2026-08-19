# syntax=docker/dockerfile:1

# ---- build stage --------------------------------------------------------
FROM golang:1.23-bookworm AS builder

WORKDIR /src

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- runtime stage -------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY --from=builder /src/config/config.example.yaml /app/config/config.yaml
COPY --from=builder /src/proto/*.swagger.json /app/proto/
COPY --from=builder /src/migrations /app/migrations

USER nonroot:nonroot

EXPOSE 4021 9090

ENTRYPOINT ["/app/server"]
