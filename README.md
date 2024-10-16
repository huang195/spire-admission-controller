# spire-admission-controller

## Build container image

```
(cd deploy; make)
make
```

This creates a custom CA, key, and cert for the webhook service. The key and
cert are then put into the webhook container image.

## Push container image

```
make push
```

## Deploy

```
cd deploy
kubectl apply -f deploy.yaml
export CA_BUNDLE=`cat ca.crt | base64 | tr -d '\n'`
cat webhook.yaml | envsubst | kubectl apply -f -)
```

This creates the mutating webhook service that the apiserver will call whenever new deployments are created.
The custom CA we created above is given to the apiserver so it can validate webhook's cert when making a TLS connection.

## Usage:

Make sure to have spire-csi-webhook already running in the cluster.

```
cd deploy
kubectl apply -f workload.yaml
```

The annotation of `spiffe.io/inject-cert: "true"` will trigger the mutating
webhook to add the spire csi ephemeral volume to the deployment. In the pod's
`/csi-identity` directory, SPIRE identities of this deployment can be found.
