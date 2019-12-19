TESTS ?= .
KUBECONFIG = ${CURDIR}/kind.kubeconfig
NAMESPACE ?= jindra-tests

kind:
	curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.6.1/kind-$$(uname)-amd64
	chmod +x ./kind

${KUBECONFIG}: kind
		docker info
		./kind create cluster --name jindra --kubeconfig ${KUBECONFIG}
		printf "waiting for node to become ready "; \
		for _ in $$(seq 0 120); \
		do  \
			sleep 1; \
			printf "."; \
			KUBECONFIG=${KUBECONFIG} kubectl get nodes|grep " Ready " && break; \
		done \

.PHONY: testreport.html
testreport.html:
	which go-junit-report
	set -o pipefail; \
	  KUBECONFIG=${KUBECONFIG} NAMESPACE=${NAMESPACE} CONCURRENT=${CONCURRENT} TESTS=${TESTS} make test | tee testreport.log; \
		res=$$?; \
		cat testreport.log | go-junit-report > testreport.xml; \
		./junit2html testreport.xml > testreport.html; \
		kubectl --kubeconfig=${KUBECONFIG} delete --wait --filename deploy.yaml
		test $$res -eq 0

test: ${KUBECONFIG}
	KUBECONFIG=${KUBECONFIG} VERBOSE=${VERBOSE} CONCURRENT=${CONCURRENT} TESTS=${TESTS} ./tests.sh

