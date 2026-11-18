# syntax=docker/dockerfile:1

ARG GO_IMAGE=golang:1.26-alpine@sha256:ecb67f987aa1a57794c91d19ade7ae0209f2f0f5496aaef9e6ae85ddf92d02ab
ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian13:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931

FROM ${GO_IMAGE} AS build
WORKDIR /src

ENV CGO_ENABLED=0 GOFLAGS=-mod=readonly

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/kwach ./cmd/kwach

FROM ${RUNTIME_IMAGE}

LABEL org.opencontainers.image.title="kwach" \
      org.opencontainers.image.description="Spoken roast generator" \
      org.opencontainers.image.licenses="MIT"

COPY --from=build --chown=65532:65532 /out/kwach /usr/local/bin/kwach

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/kwach"]
CMD ["serve"]
