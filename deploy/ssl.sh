#!/bin/bash

export APP="${1}"
export NAMESPACE="${2}"
export CSR_NAME="${APP}.${NAMESPACE}.svc"

echo "... creating ca.key"
openssl genrsa -out ca.key 4096

echo "... creating ca.crt"
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/C=US/ST=CA/L=San Francisco/O=Custom CA/OU=IT/CN=custom-ca"

echo "... creating ${APP}.key"
openssl genrsa -out ${APP}.key 2048

echo "... creating ${APP}.csr"
openssl req -new -key ${APP}.key -out ${APP}.csr -subj "/CN=${APP}.${NAMESPACE}.svc"

cat >extfile.conf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[v3_ca]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${APP}
DNS.2 = ${APP}.${NAMESPACE}
DNS.3 = ${CSR_NAME}
DNS.4 = ${CSR_NAME}.cluster.local
EOF

echo "... creating ${APP}.crt"
openssl x509 -req -in ${APP}.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out ${APP}.pem -days 3650 -sha256 -extfile extfile.conf -extensions v3_ca

