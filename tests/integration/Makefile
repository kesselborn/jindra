TESTS ?= .
KUBECONFIG = ${CURDIR}/kind.kubeconfig
CLUSTER_NAME = jindra

TEST_ARGS = KUBECONFIG=${KUBECONFIG}                                       \
            CONCURRENT=${CONCURRENT}                                       \
            TESTS=${TESTS}                                                 \
            INCLUDES=${INCLUDES}                                           \
            NO_ANSI_COLOR=${NO_ANSI_COLOR}                                 \
            FAIL_FAST=${FAIL_FAST}

default: testreport.html

kind:
	curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.6.1/kind-$$(uname)-amd64
	chmod +x ./kind

jindra-kind-cluster: ${KUBECONFIG}

.PHONY: ${KUBECONFIG}
${KUBECONFIG}: kind
		docker info
		di=$$(docker ps -q -a --filter name=${CLUSTER_NAME}-control-plane) && \
			 docker start $${di} || \
			./kind create cluster --name ${CLUSTER_NAME} --kubeconfig ${KUBECONFIG}

		printf "waiting for node to become ready ";                           \
		for _ in $$(seq 0 120);                                               \
		do                                                                    \
			sleep 1;                                                            \
			printf ".";                                                         \
			KUBECONFIG=${KUBECONFIG} kubectl get nodes|grep " Ready " && break; \
		done

.PHONY: testreport.html
testreport.html:
	which go-junit-report
	set -o pipefail;                                                           \
	  ${TEST_ARGS} make test | tee testreport.log;                             \
		res=$$?;                                                                 \
		cat testreport.log | go-junit-report > testreport.xml;                   \
		./junit2html testreport.xml > testreport.html;                           \
		kubectl --kubeconfig=${KUBECONFIG} delete --wait --filename deploy.yaml; \
		test $$res -eq 0

.manager-image: ${shell find ../.. -name "*.go"}
	make -C ../.. docker-build
	touch $@

.PHONY: load-jindra-images
load-jindra-images: .manager-image
	@for i in jindra/jindra $$(cd ../..; ls Dockerfile.*|cut -f2 -d.|sed 's/^/jindra\//') $$(cat *.yaml|grep " *image:"|cut -f2- -d:|tr -d " "|uniq); \
	do                                                                              \
		CLUSTER_NAME=${CLUSTER_NAME} IMAGE=$$i make load-image;                       \
	done

load-image:
	@kind load docker-image --name ${CLUSTER_NAME} --nodes ${CLUSTER_NAME}-control-plane $${IMAGE:?}

test: load-jindra-images
	${TEST_ARGS} ./tests.sh

