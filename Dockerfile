FROM golang:1.25-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG GIT_SHA=dev
RUN CGO_ENABLED=0 go build \
	-ldflags "-X github.com/simbachu/twisky/internal/version.BuildID=${GIT_SHA}" \
	-o /out/server \
	./cmd/server

FROM debian:bookworm-slim AS runtime

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
		ca-certificates \
		wget \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/server /app/server

ENV TWISKY_ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["/app/server"]
