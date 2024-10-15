FROM golang:1-alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o spiffe-csi-webhook ./cmd/spiffe-csi-webhook

FROM alpine:latest
RUN mkdir /ssl
COPY --from=builder /build/spiffe-csi-webhook /bin/spiffe-csi-webhook
COPY deploy/spiffe-csi-webhook.pem /ssl/spiffe-csi-webhook.pem
COPY deploy/spiffe-csi-webhook.key /ssl/spiffe-csi-webhook.key

ENTRYPOINT ["/bin/spiffe-csi-webhook"]
