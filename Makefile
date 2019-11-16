DOCKER_IMAGE=jindra/jindra
GO_FILES=${shell find pkg/controller -name "*.go"} ${shell find pkg/apis -name "*.go"} ${shell find cmd/ -name "*.go"}
PLAYGROUND_DIR=playground

all: jindra-cli k8s-pod-watcher kubectl-podstatus crij build/_output/bin/jindra

crij: pkg/jindra/tools/crij/*.go pkg/jindra/tools/crij/cmd/crij/*.go
	cd ${shell dirname $<} && go build ./cmd/$@ && cp $@ $(CURDIR)/$@

kubectl-podstatus: pkg/jindra/tools/k8spodstatus/*.go pkg/jindra/tools/k8spodstatus/cmd/kubectl-podstatus/*.go 
	cd ${shell dirname $<} && go build ./cmd/$@ && cp $@ $(CURDIR)/$@

k8s-pod-watcher: pkg/jindra/tools/k8spodstatus/*.go pkg/jindra/tools/k8spodstatus/cmd/k8s-pod-watcher/*.go 
	cd ${shell dirname $<} && go build ./cmd/$@ && cp $@ $(CURDIR)/$@

jindra-cli: pkg/jindra/*.go pkg/jindra/cmd/jindra-cli/*.go
	cd ${shell dirname $<} && go build ./cmd/$@ && cp $@ $(CURDIR)/$@

test:
	cd pkg/jindra/tools/crij && go test -v 
	cd pkg/jindra && go test -timeout 30s -v || { test $$? = 1 && code -d /tmp/expected /tmp/got; }

local: jindra-controller-image
	- kubectl -n $${NAMESPACE:?please set \$$NAMESPACE env var} create -f deploy/crds/jindra.io_jindrapipelines_crd.yaml
	ENABLE_WEBHOOKS=false OPERATOR_NAME=jindra operator-sdk up local --namespace=jindra

remote: jindra-controller-image
	test -n "$${NAMESPACE:?please set \$$NAMESPACE env var}"
	- test -z "${NO_DELETE}" && ls deploy/*.yaml|xargs -n1 kubectl -n ${NAMESPACE} delete -f || true

	- kubectl -n ${NAMESPACE} replace -f deploy/crds/jindra.io_jindrapipelines_crd.yaml || \
		kubectl -n ${NAMESPACE} create  -f deploy/crds/jindra.io_jindrapipelines_crd.yaml
	sed -i "" 's|namespace: .*$$|namespace: ${NAMESPACE}|g' deploy/cluster_role_binding.yaml
	ls deploy/*.yaml  |xargs -n1 kubectl -n ${NAMESPACE} apply  --wait -f

clean:
	- kubectl delete crd jindrapipelines.jindra.io
	- ls deploy/*.yaml|xargs -n1 kubectl -n jindra delete -f
	rm -rf build/_output/bin/jindra

deploy/crds/jindra.io_jindrapipelines_crd.yaml: ${GO_FILES}
	operator-sdk generate k8s
	operator-sdk generate openapi
	- kubectl delete crd jindrapipelines.jindra.io
	kubectl create -f deploy/crds/jindra.io_jindrapipelines_crd.yaml

build/_output/bin/jindra: deploy/crds/jindra.io_jindrapipelines_crd.yaml deploy/crds/*.yaml ${GO_FILES}
	go mod vendor
	operator-sdk build ${DOCKER_IMAGE}

jindra-controller-image: build/_output/bin/jindra
	docker push jindra/jindra
	sed -i "" 's|REPLACE_IMAGE|${DOCKER_IMAGE}|g' deploy/operator.yaml
	touch $@

update-helper-images: tools-image rsync-image runner-image pod-watcher-image

rsync-image:
	docker build -t jindra/rsync-server:latest -f Dockerfile.rsync-server .
	docker push jindra/rsync-server:latest

tools-image:
	docker build -t jindra/tools:latest -f Dockerfile.jindra-tools .
	docker push jindra/tools:latest

runner-image:
	docker build -t jindra/jindra-runner:latest -f Dockerfile.jindra-runner .
	docker push jindra/jindra-runner:latest

pod-watcher-image:
	docker build -t jindra/pod-watcher:latest -f Dockerfile.k8s-pod-watcher .
	docker push jindra/pod-watcher:latest
