#!/usr/bin/env bash

# genkeys.sh
#
# Generate a (self-signed) CA certificate along with a certificate and private key to
# be used by the image-cloner server. The certificate will be issued for the
# Common Name (CN) of `image-cloner.default.svc`, which is the
# cluster-internal DNS name for the service.

# stops the execution if a command or pipeline has an error
set -eu

echo ""
echo "[info] Generating the CA private key and cert file."
cfssl gencert -initca ca-csr.json | cfssljson -bare ca

echo ""
echo "[info] Generating the image-cloner server SSL certificate."
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -hostname="image-cloner,image-cloner.default.svc.cluster.local,image-cloner.default.svc,localhost,127.0.0.1" \
  -profile=default \
  ca-csr.json | cfssljson -bare image-cloner

echo ""
echo "[info] Creating a secret for image cloner"
cat <<EOF > ../deploy/image-cloner-tls.yaml
apiVersion: v1
kind: Secret
metadata:
  name: image-cloner-tls
type: kubernetes.io/tls
data:
  tls.crt: $(cat image-cloner.pem | base64 | tr -d '\n')
  tls.key: $(cat image-cloner-key.pem | base64 | tr -d '\n') 
EOF

if [ $? -eq 1 ]; then
      echo "[error] Error creating secret."
      exit 1
else 
      echo "[done]"
fi

echo ""
echo "[info] Generating and injecting the CA bundle into webhook configuation template."
ca_pem_b64="$(openssl base64 -A <"ca.pem")"
sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"../deploy/webhook-template.yaml" \
    > ../deploy/image-cloner-webhook.yaml

if [ $? -eq 1 ]; then
      echo "[error] Error generating and injecting the CA bundle."
      exit 1
else 
      echo "[done]"
fi
