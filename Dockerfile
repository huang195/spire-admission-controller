FROM golang:1-alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o spiffe-csi-webhook ./cmd/spiffe-csi-webhook

FROM alpine:latest
RUN mkdir /ssl
COPY --from=builder /build/spiffe-csi-webhook /bin/spiffe-csi-webhook
COPY deploy/spire-spiffe-csi-webhook.pem /ssl/spire-spiffe-csi-webhook.pem
COPY deploy/spire-spiffe-csi-webhook.key /ssl/spire-spiffe-csi-webhook.key

ENTRYPOINT ["/bin/spiffe-csi-webhook"]
