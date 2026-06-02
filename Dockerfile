# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	go build -trimpath -ldflags="-s -w" -o /out/notly-api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget \
	&& addgroup -S notly \
	&& adduser -S -G notly -h /app notly

WORKDIR /app

COPY --from=build /out/notly-api /app/notly-api

ENV API_PORT=8080

EXPOSE 8080

USER notly

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
	CMD wget -qO- "http://127.0.0.1:${API_PORT}/health" >/dev/null || exit 1

ENTRYPOINT ["/app/notly-api"]
