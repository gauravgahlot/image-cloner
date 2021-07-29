#!/usr/bin/env bash
# Copyright 2021 The image-cloner Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


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
echo "[info] Creating a TLS secret definition for image cloner"
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
