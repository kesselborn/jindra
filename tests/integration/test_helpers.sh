function wait_for_container() {
  local pipeline=$1
  local build_no=$2
  local stage=$3
  local container=$4

  for _ in $(seq 1 ${TRIES})
  do
    local pod=$(${kubectl} get pods --no-headers -l jindra.io/run=${build_no},jindra.io/pipeline=${pipeline},jindra.io/stage=${stage}|cut -f1 -d" ")
    test -n "${pod}" && break
    sleep 2
  done

  for _ in $(seq 1 ${TRIES})
  do
    local res=$(${kubectl} get pods ${pod} -ojson)
    local readyContainer=$(echo "${res}"|jq ".status.containerStatuses|map(select(.name == \"${container}\" and .ready == true))")
    test "${readyContainer}" != "[]" && break
  done

  test -n "${pod}" && echo "${pod}"
}

function wait_for_jindra_operator() {
  if [ -n "${OPERATOR_POD}" ]
  then
    echo "${OPERATOR_POD}"
    return
  fi

  if [ -z "${FIRST_RUN}" ]
  then
    NO_DELETE=1 NAMESPACE=${NAMESPACE} make -C ../ remote >&2
  fi

  for _ in $(seq 1 ${TRIES})
  do
    local pod=$(${kubectl} get pods --no-headers -l jindra-component=operator -l name=jindra|cut -f1 -d" ")
    test -n "${pod}" && break
    sleep 2
  done

  for _ in $(seq 1 ${TRIES})
  do
    sleep 2
    ${kubectl} logs ${pod}|grep "starting the webhook server." > /dev/null && break
  done

  test -n "${pod}" && echo "${pod}"
}