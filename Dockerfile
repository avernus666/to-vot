FROM golang:1.23-alpine AS build

WORKDIR /src/averna-alpha/apps/api

COPY averna-alpha/apps/api/ ./

RUN go mod tidy && go test ./... && \
    mkdir -p /out && \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/averna-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/averna-api /averna-api

EXPOSE 10000
USER nonroot:nonroot
ENTRYPOINT ["/averna-api"]
