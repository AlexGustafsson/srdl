FROM golang:1.23 AS builder

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 go build -a -ldflags="-s -w" -o srdl cmd/srdl/*.go && \
  CGO_ENABLED=0 go build -a -ldflags="-s -w" -o srdl-sub cmd/srdl-sub/*.go

FROM scratch AS export

COPY --from=builder /src/srdl srdl
COPY --from=builder /src/srdl-sub srdl-sub

FROM export

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV PATH=/

ENTRYPOINT ["srdl-sub"]
