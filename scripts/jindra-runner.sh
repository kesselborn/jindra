#!/bin/bash -e
function get_logs() {
  local pod=$1

  for container in \
       $(kubectl get pod ${pod} -ojson |jq -r ".spec.initContainers[]|.name // empty") \
       $(kubectl get pod ${pod} -ojson |jq -r ".spec.containers[]|.name // empty")
  do
    echo -e "\n\n\n################################################################ start: logs for ${pod}/${container}"
    kubectl logs ${pod} -c ${container}
    echo          "################################################################ end  : logs for ${pod}/${container}"
  done
  kubectl get pod ${pod} -ojson|jq .status
}

function wait_for_containers() {
  local pod=$1
  local containers=$2

  printf "waiting for pod %s/%s " ${pod} ${containers}
  while true
  do
    local state=$(wget -qO- localhost:8080/pod/${pod}?containers=${containers})
    printf "."
    test  "${state}" = "Completed" && { echo; local return_value=0; }
    test  "${state}" = "Failed"    && { echo; local return_value=1; }

    test -n "${return_value}" && return ${return_value}
    sleep 3
  done
}

block_until_finished() {
  local name=$1
  local wait_for=$(kubectl get pod ${name} -ojson |jq -r ".metadata.annotations[\"${WAIT_FOR_ANNOTATION_KEY}\"] // empty")
  local outputs=$(kubectl get pod ${name} -ojson |jq -r ".metadata.annotations[\"${OUT_RESOURCE_ANNOTATION_KEY}\"] // empty")

  local outputs_container_names=""
  for c in $(echo "${outputs}"|tr "," " ")
  do
    if [ -n "${outputs_container_names}" ]
    then
      outputs_container_names="${outputs_container_names},"
    fi
    outputs_container_names="${outputs_container_names}${OUT_RESOURCE_CONTAINER_NAME_PREFIX}${c}"
  done

  if wait_for_containers ${name} ${wait_for}
  then
    if wait_for_containers ${name} ${outputs_container_names}
    then
      return 0
    else
      echo "waiting for outputs ${name}/${output} failed"
      return 1
    fi
  else
    echo "waiting for steps ${name}/${wait_for} failed"
    return 2
  fi
}

run_pod() {
  local config_file=$1
  test -e ${config_file} || return 0

  # TODO: look for a better way for this
  local name=$(eval "cat<<-EOF
$(<${config_file})
EOF
  " | kubectl apply -f-|grep -o "pod/[^ ][^ ]*"|sed 's#pod/##')

  kubectl patch pod ${name} --patch "$(cat /tmp/patch.yaml)"
  block_until_finished ${name}
  local pod_res=$?
  get_logs ${name}
    kubectl delete --wait=false pod ${name}
  if [ "${pod_res}" = "0" ]
  then
    return 0
  else
    kubectl delete --wait=true pod ${name}
    return 1
  fi
}

cat<<EOF >/tmp/patch.yaml
metadata:
  labels:
    ${RUN_LABEL_KEY}: "${JINDRA_PIPELINE_RUN_NO}"
    ${PIPELINE_LABEL_KEY}: "${JINDRA_PIPELINE_NAME}"
EOF

kubectl patch pod ${MY_NAME} --patch "$(cat /tmp/patch.yaml)"
kubectl patch job jindra.${JINDRA_PIPELINE_NAME}.${JINDRA_PIPELINE_RUN_NO} --patch "$(cat /tmp/patch.yaml)"

cat<<EOF >>/tmp/patch.yaml
  ownerReferences:
  - apiVersion: batch/v1
    controller: true
    kind: Job
    name: ${MY_NAME}
    uid: ${MY_UID}
EOF

(set -x;
kubectl patch secret    $(printf "${RSYNC_KEY_NAME_FORMAT_STRING}"  "${JINDRA_PIPELINE_NAME}" ${JINDRA_PIPELINE_RUN_NO}) --patch "$(cat /tmp/patch.yaml)"
kubectl patch configmap $(printf "${CONFIG_MAP_NAME_FORMAT_STRING}" "${JINDRA_PIPELINE_NAME}" ${JINDRA_PIPELINE_RUN_NO}) --patch "$(cat /tmp/patch.yaml)"
)

for f in $(ls /jindra/stages/*.yaml|grep -v "[0-9][0-9]-on-error.yaml$"|grep -v "[0-9][0-9]-on-success.yaml$"|grep -v "[0-9][0-9]-final.yaml$")
do
  run_pod $f
  res=$?
  if [ "${res}" != "0" ]
  then
    break
  fi
done

if [ "${res}" = "0" ]
then
  run_pod /jindra/stages/[0-9][0-9]-on-success.yaml
else
  run_pod /jindra/stages/[0-9][0-9]-on-error.yaml
fi

run_pod /jindra/stages/[0-9][0-9]-final.yaml

rm /jindra/semaphores/stages-running
exit ${res}
