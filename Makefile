DOCKER_IMAGE=jindra/jindra
GO_FILES=${shell find pkg/controller -name "*.go"} ${shell find pkg/apis -name "*.go"}

all: k8s-pod-watcher kubectl-podstatus crij build/_output/bin/jindra

crij: pkg/jindra/tools/crij/*.go pkg/jindra/tools/crij/cmd/crij/*.go
	cd ${shell dirname $<} && go build ./cmd/$@ && cp $@ $(CURDIR)/$@

kubectl-podstatus: pkg/jindra/tools/k8spodstatus/*.go pkg/jindra/tools/k8spodstatus/cmd/kubectl-podstatus/*.go 
	cd ${shell dirname $<} && go build ./cmd/$@ && cp $@ $(CURDIR)/$@

k8s-pod-watcher: pkg/jindra/tools/k8spodstatus/*.go pkg/jindra/tools/k8spodstatus/cmd/k8s-pod-watcher/*.go 
	cd ${shell dirname $<} && go build ./cmd/$@ && cp $@ $(CURDIR)/$@

test:
	cd pkg/jindra/tools/crij && go test -v 
	cd pkg/jindra && go test -timeout 30s -v || { test $$? = 1 && code -d /tmp/expected /tmp/got; }

local: docker-image
	- kubectl create -f deploy/crds/jindra_v1alpha1_jindrapipeline_crd.yaml
	OPERATOR_NAME=jindra operator-sdk up local --namespace=jindra

clean:
	- kubectl delete crd jindrapipelines.jindra.io
	- ls deploy/*.yaml|xargs -n1 kubectl -n jindra delete -f
	rm -rf build/_output/bin/jindra

remote: docker-image
	- kubectl create -f deploy/crds/jindra_v1alpha1_jindrapipeline_crd.yaml
	- ls deploy/*.yaml|xargs -n1 kubectl -n jindra delete -f
	ls deploy/*.yaml|xargs -n1 kubectl -n jindra create -f

deploy/crds/jindra_v1alpha1_jindrapipeline_crd.yaml: ${GO_FILES}
	operator-sdk generate k8s
	operator-sdk generate openapi
	- kubectl delete crd jindrapipelines.jindra.io
	kubectl create -f deploy/crds/jindra_v1alpha1_jindrapipeline_crd.yaml

build/_output/bin/jindra: deploy/crds/*.yaml ${GO_FILES}
	operator-sdk build ${DOCKER_IMAGE}

docker-image: build/_output/bin/jindra
	docker push jindra/jindra
	sed -i "" 's|REPLACE_IMAGE|${DOCKER_IMAGE}|g' deploy/operator.yaml
	touch $@

tools-image:
	docker build -t jindra/tools:latest -f Dockerfile.jindra-tools .
	docker push jindra/tools:latest

runner-image:
	docker build -t jindra/jindra-runner:latest -f Dockerfile.jindra-runner .
	docker push jindra/jindra-runner:latest

pod-watcher-image:
	docker build -t jindra/pod-watcher:latest -f Dockerfile.k8s-pod-watcher .
	docker push jindra/pod-watcher:latest
