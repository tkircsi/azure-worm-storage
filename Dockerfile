############################
# STEP 1 build executable binary
############################
FROM --platform=${BUILDPLATFORM} golang:1.15.2-alpine3.12 AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git curl
WORKDIR $GOPATH/src/app
ENV GO111MODULE="on"
ENV CGO_ENABLED=0
COPY . .
# Fetch dependencies.
# Using go get.
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go get -d -v && go build -o /go/bin/app
# Build the binary.
# RUN go build -o /go/bin/gorilla-rest-api
############################
# STEP 2 build a small image
############################
# scratch don't have curl. Need to make a Go binary client to check endpoint
# FROM scratch
FROM alpine:3.12

RUN apk update && apk add --no-cache curl
# Copy our static executable.
COPY --from=builder /go/bin/app /go/bin/app

EXPOSE 5000
HEALTHCHECK --interval=5s --timeout=3s --start-period=1s --retries=3 \
  CMD curl -f http://localhost:5000/healthcheck || exit 1
# Run the binary.
ENTRYPOINT ["/go/bin/app"]