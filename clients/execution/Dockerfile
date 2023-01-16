# Docker container spec for building builder.
FROM golang:1.18-alpine as builder

RUN apk add --no-cache gcc musl-dev linux-headers git

ARG branch=main
ARG repo=https://github.com/flashbots/builder.git
RUN git clone --depth 1 --branch $branch $repo /go-ethereum

RUN cd /go-ethereum && go run build/ci.go install -static ./cmd/geth

# Pull Geth into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

# Export the usual networking ports to allow outside access to the node
EXPOSE 8545 8546 8547 8551 6060 30303 30303/udp

ENTRYPOINT ["geth"]
