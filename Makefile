all:
	podman build -t haih/spiffe-csi-webhook:latest .

push:
	podman push haih/spiffe-csi-webhook:latest
