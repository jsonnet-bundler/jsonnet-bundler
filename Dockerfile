###########
# Builder #
###########

FROM golang:1.14-alpine AS builder

# Switch to Go modules directory
WORKDIR /go/src/github.com/jsonnet-bundler/jsonnet-bundler

# Install build tools
RUN apk add --no-cache alpine-sdk bash

# Copy source directory
COPY . .

# Build static binary against Git version
RUN make static

########
# Prod #
########

# Production image
FROM alpine:3.12 AS production

# Copy static binary from 'builder'
COPY --from=builder /go/src/github.com/jsonnet-bundler/jsonnet-bundler/_output/jb /usr/local/bin/

# Set image entrypoint
ENTRYPOINT ["/usr/local/bin/jb"]
