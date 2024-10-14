all:
	podman build -t haih/spire-spiffe-csi-webhook:latest .

push:
	podman push haih/spire-spiffe-csi-webhook:latest
