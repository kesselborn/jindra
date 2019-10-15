DOCKER_IMAGE=jindra/jindra
GO_FILES=${shell find . -name "*.go"}

build: build/_output/bin/jindra

test:
	cd pkg/jindra && go test -timeout 30s -v || { test $$? = 1 && vimdiff /tmp/expected /tmp/got; }

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

