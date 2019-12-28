# Image URL to use all building/pushing image targets
IMG ?= jindra/jindra:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
GO_FILES = $(shell find . -name "*.go")
NAMESPACE ?= jindra
DEPLOY_CONFIG ?= deploy.yaml
TESTS ?= .

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager bin/kubectl-podstatus bin/k8s-pod-watcher bin/jindra-cli bin/crij

bin/crij bin/kubectl-podstatus bin/k8s-pod-watcher bin/jindra-cli: ${GO_FILES}
	go build -o $@ ./cmd/$$(basename $@)

# Run tests
unittests: generate fmt vet manifests
	rm -rf /tmp/exected /tmp/got
	go test -run ${TESTS} -v ./api/... -coverprofile cover.out || { test $$? -eq 1 -a -e /tmp/expected && code -d /tmp/expected /tmp/got; exit 1; }
	go test -run ${TESTS} -v ./crij/... -coverprofile cover.out
	go test -run ${TESTS} -v ./k8spodstatus/... -coverprofile cover.out

test: unittests
	go test -run ${TESTS} -v ./controllers/... -coverprofile cover.out || { test $$? = 1 -a -e /tmp/expected && code -d /tmp/expected /tmp/got; }

# Build manager binary
manager: bin/manager

bin/manager: manifests generate fmt vet ${GO_FILES}
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	- KUBECONFIG=${KUBECONFIG} kubectl delete crd pipelines.ci.jindra.io
	kustomize build config/crd | KUBECONFIG=${KUBECONFIG} kubectl create -f -

reset-config:
	cd config/manager && kustomize edit set image controller=${IMG}
	cd config/default && kustomize edit set namespace jindra
	cd config/default && kustomize edit set nameprefix ""


# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
${DEPLOY_CONFIG}: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	cd config/default && kustomize edit set namespace ${PREFIX}${NAMESPACE}
	cd config/default && kustomize edit set nameprefix "${PREFIX}"

	KUBECONFIG=${KUBECONFIG} kubectl kustomize config/$${KUSTOMIZE_CONFIG:-default} > $@
	sed -i "" -e 's/name: ${PREFIX}system$$/name: ${PREFIX}${NAMESPACE}/' \
		        -e 's/- psp-jindra-controller$$/- ${PREFIX}psp-jindra-controller/' \
						$@

deploy: ${DEPLOY_CONFIG}
	KUBECONFIG=${KUBECONFIG} kubectl -n ${PREFIX}${NAMESPACE} apply -f $<

.PHONY: config/webhook-certs/cert-secret.yaml
config/webhook-certs/cert-secret.yaml:
	SERVICE_NAME=${PREFIX}webhook-service NAMESPACE=${PREFIX}${NAMESPACE} ./jindra-pki.sh > $@

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen config/webhook-certs/cert-secret.yaml
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: unittests
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.2 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

tools-images: jindra-runner tools pod-watcher rsync-server

jindra-runner tools pod-watcher rsync-server:
	docker build -t jindra/$@ -f Dockerfile.$@ .
	docker push     jindra/$@
