FROM golang:1.23-alpine AS build
WORKDIR /src/apps/api
COPY apps/api/go.mod ./
RUN go mod download
WORKDIR /src
COPY apps/api ./apps/api
WORKDIR /src/apps/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/averna-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/averna-api /averna-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/averna-api"]
