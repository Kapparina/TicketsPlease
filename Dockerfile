FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG GIT_TAG=unknown
ARG COMMIT=unknown

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build \
      -trimpath \
      -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.GitTag=${GIT_TAG}'" \
      -o /out/ticketsplease \
      github.com/kapparina/ticketsplease

FROM alpine:3.20

ARG VERSION=dev
ARG GIT_TAG=unknown
ARG COMMIT=unknown

LABEL org.opencontainers.image.title="TicketsPlease" \
      org.opencontainers.image.description="Discord ticket manager bot" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.source="https://github.com/kapparina/ticketsplease"

RUN apk add --no-cache ca-certificates \
    && addgroup -S ticketsplease \
    && adduser -S -G ticketsplease -h /nonexistent -s /sbin/nologin ticketsplease

COPY --from=build /out/ticketsplease /usr/local/bin/ticketsplease
COPY config.example.toml /config/config.toml

USER ticketsplease:ticketsplease

ENTRYPOINT ["/usr/local/bin/ticketsplease"]

CMD ["-config", "/config/config.toml"]
