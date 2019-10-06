update-stages:
	-kubectl -n jindra delete configmap jindra.http-fs-42.stages
	-kubectl -n jindra delete secret jindra.http-fs-42.rsync-keys
	kubectl -n jindra create secret generic jindra.http-fs-42.rsync-keys --from-file="priv=rsync" --from-file="pub=rsync.pub"
	kubectl -n jindra create configmap jindra.http-fs-42.stages --from-file=jindra.http-fs-42.01-build-binary.yaml=jindra.http-fs-42.01-build-binary.yaml --from-file=jindra.http-fs-42.02-build-docker-image.yaml=jindra.http-fs-42.02-build-docker-image.yaml --from-file=jindra.http-fs-42.on-failure.yaml=jindra.http-fs-42.on-failure.yaml --from-file=jindra.http-fs-42.on-success.yaml=jindra.http-fs-42.on-success.yaml

deploy: update-stages
	-kubectl -n jindra delete job jindra.http-fs-42
	kubectl -n jindra apply -f jindra-runner-permissions.yaml
	kubectl -n jindra apply -f jindra.http-fs-42.yaml

update-images: tools-image rsync-image runner-image pod-watcher-image

runner-image:
	docker build -t jindra/jindra-runner:latest -f Dockerfile.jindra-runner .
	docker push jindra/jindra-runner:latest

pod-watcher-image:
	docker build -t jindra/pod-watcher:latest -f Dockerfile.k8s-pod-watcher .
	docker push jindra/pod-watcher:latest

rsync-image:
	docker build -t jindra/rsync-server:latest -f Dockerfile.rsync-server .
	docker push jindra/rsync-server:latest

tools-image:
	docker build -t jindra/tools:latest -f Dockerfile.jindra-tools .
	docker push jindra/tools:latest
