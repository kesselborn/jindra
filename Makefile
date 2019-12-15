
# Image URL to use all building/pushing image targets
IMG ?= jindra/jindra:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
GO_FILES = $(shell find . -name "*.go")

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
test: generate fmt vet manifests
	rm -rf /tmp/exected /tmp/got
	go test -v ./... -coverprofile cover.out || { test $$? = 1 -a -e /tmp/expected && code -d /tmp/expected /tmp/got; }

# Build manager binary
manager: bin/manager

bin/manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl create -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	# double execute as apply saves the old config in an annotation which gets too big
	kubectl kustomize config/default | kubectl replace -f- || kubectl kustomize config/default | kubectl apply -f-

.pki/server.crt:
	./jindra-pki.sh

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen .pki/server.crt
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
docker-build: test
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