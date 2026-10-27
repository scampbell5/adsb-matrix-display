FROM golang:1.25-alpine AS build
RUN apk add --no-cache ca-certificates git
WORKDIR /src

# Prefer the public Go module proxy, but fall back to fetching directly from
# the module's source repository on any proxy/network error.
ENV GOPROXY=https://proxy.golang.org|direct

COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod \
    -trimpath -ldflags='-s -w' \
    -o /out/adsb-matrix-display ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/adsb-matrix-display /usr/local/bin/adsb-matrix-display
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -q -O - http://127.0.0.1:8080/healthz >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/adsb-matrix-display"]
