FROM --platform=${BUILDPLATFORM} golang:1.25 AS builder

WORKDIR /src

# Use the toolchain specified in go.mod, or newer
ENV GOTOOLCHAIN=auto

COPY go.mod go.sum .
RUN go mod download && go mod verify

COPY cmd cmd
COPY internal internal

ARG TARGETARCH
RUN GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -a -ldflags="-s -w" -o srdl cmd/srdl/*.go && \
  GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -a -ldflags="-s -w" -o srdl-sub cmd/srdl-sub/*.go

FROM scratch AS export

COPY --from=builder /src/srdl srdl
COPY --from=builder /src/srdl-sub srdl-sub

FROM export

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV PATH=/

ENTRYPOINT ["srdl-sub"]
