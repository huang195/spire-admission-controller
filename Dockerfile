FROM golang:1-alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o spiffe-csi-webhook ./cmd/spiffe-csi-webhook

FROM alpine:latest
COPY --from=builder /build/spiffe-csi-webhook /bin/spiffe-csi-webhook

ENTRYPOINT ["/bin/spiffe-csi-webhook"]
