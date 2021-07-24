FROM debian

WORKDIR /tls

RUN apt-get update && apt-get install -y curl && \
  curl -L https://github.com/cloudflare/cfssl/releases/download/v1.5.0/cfssl_1.5.0_linux_amd64 -o /usr/local/bin/cfssl && \
  curl -L https://github.com/cloudflare/cfssl/releases/download/v1.5.0/cfssljson_1.5.0_linux_amd64 -o /usr/local/bin/cfssljson && \
  chmod +x /usr/local/bin/cfssl && \
  chmod +x /usr/local/bin/cfssljson

USER 1000

ENTRYPOINT [ "./gencerts.sh" ]
