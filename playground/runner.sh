#!/bin/bash
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
  local wait_for=$(kubectl get pod ${name} -ojson |jq -r '.metadata.annotations["jindra.io/wait-for"] // empty')
  local outputs=$(kubectl get pod ${name} -ojson |jq -r '.metadata.annotations["jindra.io/outputs"] // empty')

  if wait_for_containers ${name} ${wait_for}
  then
    if wait_for_containers ${name} ${outputs}
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
  # TODO: look for a better way for this
  local name=$(eval "cat<<-EOF
$(<${config_file})
EOF
  " | kubectl apply -f-|grep -o "pod/[^ ][^ ]*"|sed 's#pod/##')

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

for f in $(ls /jindra/stages/*.yaml|grep -v ".on-failure.yaml$"|grep -v ".on-success.yaml$")
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
  run_pod /jindra/stages/*.on-success.yaml
else
  run_pod /jindra/stages/*.on-failure.yaml
fi

rm /jindra/semaphores/stages-running
exit ${res}
