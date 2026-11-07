FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/relay ./cmd/relay
FROM alpine:3.21
RUN adduser -D -H -u 10001 relay && mkdir -p /var/data /app/web && chown -R relay:relay /var/data /app
WORKDIR /app
COPY --from=build /out/relay ./relay
COPY web ./web
USER relay
ENV PORT=8080 DATABASE_PATH=/var/data/relay.db
EXPOSE 8080
ENTRYPOINT ["./relay"]
