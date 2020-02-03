JINDRA_CLI = ../../bin/jindra-cli

update-stages:
	-kubectl -n jindra delete configmap jindra.http-fs.42.stages
	-kubectl -n jindra delete secret jindra.http-fs.42.rsync-keys

	kubectl -n jindra create secret generic jindra.http-fs.42.rsync-keys --from-file="priv=rsync" --from-file="pub=rsync.pub"
	kubectl -n jindra create configmap jindra.http-fs.42.stages \
	 --from-file=01-build-go-binary.yaml=jindra.http-fs.42.01-build-go-binary.yaml \
	 --from-file=02-build-docker-image.yaml=jindra.http-fs.42.02-build-docker-image.yaml \
	 --from-file=03-on-success.yaml=jindra.http-fs.42.03-on-success.yaml \
	 --from-file=04-on-error.yaml=jindra.http-fs.42.04-on-error.yaml \
	 --from-file=05-final.yaml=jindra.http-fs.42.05-final.yaml

deploy: update-stages
	-kubectl -n jindra delete job jindra.http-fs.42
	kubectl -n jindra apply -f jindra-runner-permissions.yaml
	kubectl -n jindra apply -f jindra.http-fs.42.yaml
	watch kubectl -n jindra get --no-headers jobs,pods,configmaps,secrets -l jindra.io/pipeline

secrets:
	-kubectl -n jindra create secret generic deploy-key --from-file=key=jindra-deploy.key
	kubectl -n jindra apply -f dockerhub-secret.yaml

.PHONY: ${JINDRA_CLI}
 ${JINDRA_CLI}:
	make -C ../.. bin/jindra-cli

jindra.http-fs.42.01-build-go-binary.yaml: ${JINDRA_CLI}
	${JINDRA_CLI} -b 42 -c pipeline-example.yaml stage 01-build-go-binary > $@

jindra.http-fs.42.02-build-docker-image.yaml: ${JINDRA_CLI}
	${JINDRA_CLI} -b 42 -c pipeline-example.yaml stage 02-build-docker-image > $@

jindra.http-fs.42.03-on-success.yaml: ${JINDRA_CLI}
	${JINDRA_CLI} -b 42 -c pipeline-example.yaml stage 03-on-success > $@

jindra.http-fs.42.04-on-error.yaml: ${JINDRA_CLI}
	${JINDRA_CLI} -b 42 -c pipeline-example.yaml stage 04-on-error > $@

jindra.http-fs.42.05-final.yaml: ${JINDRA_CLI}
	${JINDRA_CLI} -b 42 -c pipeline-example.yaml stage 05-final > $@

jindra.http-fs.42.stages.yaml: ${JINDRA_CLI}
	${JINDRA_CLI} -b 42 -c pipeline-example.yaml configmap > $@

jindra.http-fs.42.yaml: ${JINDRA_CLI}
	${JINDRA_CLI} -b 42 -c pipeline-example.yaml runner > $@
