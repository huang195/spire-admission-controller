## Create the webhook certificate and key

```
make
```

and in the directory, there should be 2 additional files created.
`spiffe-csi-webhook.key` is the x509 key, and `spiffe-csi-webhook.pem` is the
certificate. These files are expected to be used by the upper layer Dockerfile
to be included in the webhook container image.

## Create the MutatingWebhook 

```
export CA_BUNDLE=`./get-ca-bundle.sh` 
cat webhook.yaml | envsubst | kubectl apply -f -
```

This creates the mutating webhook custom resource
