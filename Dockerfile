# Render Web Service — Docker runtime
FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache unzip
COPY averna-source.zip /tmp/averna-source.zip
RUN unzip -q /tmp/averna-source.zip -d /src
WORKDIR /src/averna-source/apps/api
# Generate go.sum from go.mod in the build environment, then compile.
RUN go mod tidy && go test ./... && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/averna-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/averna-api /averna-api
EXPOSE 10000
USER nonroot:nonroot
ENTRYPOINT ["/averna-api"]
