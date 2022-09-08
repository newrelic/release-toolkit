# Builder for GHA does not use BUILDKIT, so we cannot use a multiarch-friendly builder image.
# FROM --platform=$BUILDPLATFORM golang:1.18-alpine as builder
FROM golang:1.19.1-alpine as builder

WORKDIR /src
COPY src/go.* ./
# Download modules on a different layer to be more cache-friendly
RUN go mod download
# This copies the _contents_ of ./src into /src.
COPY src /src
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /bin/rt ./cmd

FROM alpine as runner
COPY --from=builder /bin/rt /bin/rt
ENTRYPOINT [ "/bin/rt" ]
