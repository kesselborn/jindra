#!/bin/sh

kubectl delete configmap jindra-tools
kubectl create configmap jindra-tools --from-file=env-to-concourse-resources-json.sh=env-to-concourse-resources-json.sh
