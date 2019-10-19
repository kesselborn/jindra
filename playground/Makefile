update-stages:
	-kubectl -n jindra delete configmap jindra.http-fs.42.stages
	-kubectl -n jindra delete secret jindra.http-fs.42.rsync-keys
	kubectl -n jindra create secret generic jindra.http-fs.42.rsync-keys --from-file="priv=rsync" --from-file="pub=rsync.pub"
	kubectl -n jindra create configmap jindra.http-fs.42.stages \
	 --from-file=01-build-go-binary.yaml=jindra.http-fs.42.01-build-go-binary.yaml \
	 --from-file=02-build-docker-image.yaml=jindra.http-fs.42.02-build-docker-image.yaml \
	 --from-file=04-on-error.yaml=jindra.http-fs.42.04-on-error.yaml \
	 --from-file=03-on-success.yaml=jindra.http-fs.42.03-on-success.yaml

deploy: update-stages
	-kubectl -n jindra delete job jindra.http-fs.42
	kubectl -n jindra apply -f jindra-runner-permissions.yaml
	kubectl -n jindra apply -f jindra.http-fs.42.yaml

secrets:
	-kubectl -n jindra create secret generic deploy-key --from-file=key=jindra-deploy.key
	kubectl -n jindra apply -f dockerhub-secret.yaml

update-images: tools-image rsync-image runner-image pod-watcher-image

pod-watcher-image:
	docker build -t jindra/pod-watcher:latest -f Dockerfile.k8s-pod-watcher .
	docker push jindra/pod-watcher:latest

rsync-image:
	docker build -t jindra/rsync-server:latest -f Dockerfile.rsync-server .
	docker push jindra/rsync-server:latest
