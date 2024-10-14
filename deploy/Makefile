APP = spiffe-csi-webhook
NAMESPACE = spire-system

.PHONY:cert
cert:
	@./ssl.sh $(APP) $(NAMESPACE)

.PHONY:clean
clean:
	rm -vf *.key *.pem *.cert *.crt *.csr csr.conf
