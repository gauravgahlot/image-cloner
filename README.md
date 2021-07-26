# Image Cloner Admission Webhook

## Prerequisites

- [make][5]
- [docker][6]
- [kubectl][7]
- [minikube][8]

## make

The repository uses a Makefile for different operations.
Here are the available targets:

```sh
$ make
build    build Docker image for image cloner
deploy   register and deploy webhook in K8s cluster
gen      generate certificates, K8s TLS secret, and webhook configuration
help     print this help
lint     run lint and go mod tidy
test     run tests
```

## Docker Registry Authentication

For image cloner to be able to push to a Docker registry, it requires user
credentials for authentication. These credentials are made available to the
application by mounting `registry-auth` secret as a volume. 

You can create create the `registry-auth` secret using below command:

```sh
kubectl create secret generic registry-auth \
  --from-literal=username=<REGISTRY-USERNAME> \
  --from-literal=password=<REGISTRY-PASSWORD>
```

## TLS Certificates

The common name (CN) of the certificate must match the server name used by the
Kubernetes API server, which for internal services is
`<service-name>.<namespace>.svc`, i.e., `image-cloner.default.svc`
in our case.

For generating the TLS Certificates we have the [gencerts.sh][1] script. The
script:
- generates the certificates.
- creates `deploy/image-cloner-tls.yaml` definition for TLS secret with server
certificate and private key. ([ref][2])
- creates `deploy/image-cloner-webhook.yaml` definition for webhook
configuration by updating the `tls/webhook-template.yaml` with CA bundle. 
([ref][3])

Let's generate the certificates using the command:

```sh
$ make gen
docker build -t generate-certs:v1 ./tls
Sending build context to Docker daemon  9.216kB
Step 1/5 : FROM debian
 ---> 0980b84bde89
Step 2/5 : WORKDIR /tls
 ---> Using cache
 ---> 865d213526c4
Step 3/5 : RUN apt-get update && apt-get install -y curl &&   curl -L https://github.com/cloudflare/cfssl/releases/download/v1.5.0/cfssl_1.5.0_linux_amd64 -o /usr/local/bin/cfssl &&   curl -L https://github.com/cloudflare/cfssl/releases/download/v1.5.0/cfssljson_1.5.0_linux_amd64 -o /usr/local/bin/cfssljson &&   chmod +x /usr/local/bin/cfssl &&   chmod +x /usr/local/bin/cfssljson
 ---> Using cache
 ---> f58fc6152f41
Step 4/5 : USER 1000
 ---> Using cache
 ---> e3851aef2574
Step 5/5 : ENTRYPOINT [ "./gencerts.sh" ]
 ---> Using cache
 ---> 96bb21095120
Successfully built 96bb21095120
Successfully tagged generate-certs:v1
docker run --rm -it -v /home/gg/go/src/github.com/gauravgahlot/image-cloner/tls:/tls -v /home/gg/go/src/github.com/gauravgahlot/image-cloner/deploy:/deploy generate-certs:v1

[info] Generating the CA private key and cert file.
2021/07/24 11:33:07 [INFO] generating a new CA key and certificate from CSR
2021/07/24 11:33:07 [INFO] generate received request
2021/07/24 11:33:07 [INFO] received CSR
2021/07/24 11:33:07 [INFO] generating key: rsa-2048
2021/07/24 11:33:07 [INFO] encoded CSR
2021/07/24 11:33:07 [INFO] signed certificate with serial number 331726822647713658807283560194435768294244203408

[info] Generating the image-cloner server SSL certificate.
2021/07/24 11:33:07 [INFO] generate received request
2021/07/24 11:33:07 [INFO] received CSR
2021/07/24 11:33:07 [INFO] generating key: rsa-2048
2021/07/24 11:33:07 [INFO] encoded CSR
2021/07/24 11:33:07 [INFO] signed certificate with serial number 109654362043423132224281138530656190532879658577

[info] Creating a secret for image cloner
[done]

[info] Generating and injecting the CA bundle into webhook configuation template.
[done]
```

## Docker Image for Image Cloner

- Build the Docker image for `image-cloner` using the command:

```sh
make build
```

- The build creates the `image-cloner:v1` image.
- You can now tag and push the image to a registry of your choice. 
- You will also need to update the Docker image in the [deployment][4] file.

## Deploy Image Cloner

Once the image has been pushed to a registry or loaded into Kind, we can then
deploy image cloner using the command:

```sh
$ make deploy
kubectl apply -f deploy/image-cloner-tls.yaml \
  -f deploy/image-cloner-svc.yaml \
  -f deploy/image-cloner-deploy.yaml \
  -f deploy/image-cloner-webhook.yaml
secret/image-cloner-tls created
service/image-cloner created
deployment.apps/image-cloner created
mutatingwebhookconfiguration.admissionregistration.k8s.io/image-cloner created
```

In order to test if the server is up and running, we can forward a local port 
to the service:

```sh
kubectl port-forward service/image-cloner 8443:443 &
```

We should now be able to send a request to the server:

```sh
curl -k https://localhost:8443/readz
```

If you receive `OK` in response, then the server is up and running.

## make test

There are a few unit tests available to test the solution. The current status is:

```sh
$ make test
go clean -testcache
go test ./... -coverprofile coverage.out && go tool cover -func coverage.out
?   	github.com/gauravgahlot/image-cloner	[no test files]
?   	github.com/gauravgahlot/image-cloner/internal/docker	[no test files]
ok  	github.com/gauravgahlot/image-cloner/internal/server	0.008s	coverage: 74.1% of statements
github.com/gauravgahlot/image-cloner/internal/server/clone.go:26:	cloneImage			69.0%
github.com/gauravgahlot/image-cloner/internal/server/clone.go:74:	validateReviewRequest		71.4%
github.com/gauravgahlot/image-cloner/internal/server/clone.go:87:	writeAdmissionReviewResponse	80.0%
github.com/gauravgahlot/image-cloner/internal/server/config.go:16:	configTLS			0.0%
github.com/gauravgahlot/image-cloner/internal/server/mock.go:7:		withRegistry			100.0%
github.com/gauravgahlot/image-cloner/internal/server/mock.go:11:	withRegistryUser		100.0%
github.com/gauravgahlot/image-cloner/internal/server/mock.go:21:	ImagePull			100.0%
github.com/gauravgahlot/image-cloner/internal/server/mock.go:24:	ImagePush			100.0%
github.com/gauravgahlot/image-cloner/internal/server/mock.go:28:	ImageTag			100.0%
github.com/gauravgahlot/image-cloner/internal/server/readyz.go:9:	readyz				75.0%
github.com/gauravgahlot/image-cloner/internal/server/response.go:47:	createResponse			90.0%
github.com/gauravgahlot/image-cloner/internal/server/response.go:70:	tryCreatePatches		100.0%
github.com/gauravgahlot/image-cloner/internal/server/response.go:103:	isUsingBackupRegistry		100.0%
github.com/gauravgahlot/image-cloner/internal/server/response.go:110:	newImage			100.0%
github.com/gauravgahlot/image-cloner/internal/server/response.go:118:	createErrorResponse		100.0%
github.com/gauravgahlot/image-cloner/internal/server/server.go:24:	Setup				0.0%
github.com/gauravgahlot/image-cloner/internal/server/server.go:46:	Serve				0.0%
total:									(statements)			74.1%

```

## Troubleshooting

- Q: Why doesn't solution work on Kind?

The solution uses Docker SDK (Go) to pull/tag/push a Docker image. For that to work, we mount 
`/var/run/docker.sock` as a volume to the image cloner. A Kind cluster _does not_ use docker socket.
Instead, it uses the containerd socket available at `/var/run/containerd/containerd.sock`. Minikube,
on the other hand, uses the docker socket and that's why it has beed added as a prerequisite.

- Q: Why the image cloner pod fails to run?

Please ensure that you deploy image cloner _before_ you register the admission webhook. If
the webhook is registered first, then when you deploy image cloner the webhook will send a request
to `image-cloner` service at `/clone-image`. Since image cloner is not yet deployed, there is
no server to respond to the request and consequently the webhook will timeout. As a result,
the deployment will fail.

[1]: tls/gencerts.sh
[2]: tls/gencerts.sh#L28
[3]: tls/gencerts.sh#L48
[4]: deploy/image-cloner-deploy.yaml#L21
[5]: https://www.gnu.org/software/make/
[6]: https://docs.docker.com/get-docker/
[7]: https://kubernetes.io/docs/tasks/tools/#kubectl
[8]: https://minikube.sigs.k8s.io/docs/start/

