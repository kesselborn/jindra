#!/bin/sh

source spec.sh

NAMESPACE=jindra-tests-${RANDOM}
kubectl="kubectl -n ${NAMESPACE}"
jindra_cli=../jindra-cli

TRIES=${TRIES:-100}

function run() {
  local pipeline=$1
  ${jindra_cli} -build-no 42 -c ${pipeline} all|${kubectl} apply -f-
}

# you can put tests in other files, just include them
# for f in test_*.sh; do include $f; done

# execute commands before running any test
before_all() {
  kubectl create namespace ${NAMESPACE}
  ${kubectl} apply -f ../playground/jindra-runner-permissions.yaml
}

# execute command when all tests have beend executed
after_all() {
  kubectl delete --wait=false namespace ${NAMESPACE}
}

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
    res=$(${kubectl} get pods ${pod} -ojson|jq ".status.containerStatuses|map(select(.name == \"${container}\" and .ready == true))")
    test "${res}" != "[]" && break
  done

  test -n "${pod}" && echo "${pod}"
}

it_should_correctly_transit_files() {
  local build_no=${RANDOM}
  ${jindra_cli} -build-no ${build_no} -c transit-test.yaml all|${kubectl} apply -f-
  pod=$(wait_for_container transit-test ${build_no} three sleep)

  assert_true "test -n '${pod}'" "container three/sleep should be running"
  assert_eq "$(${kubectl} exec ${pod} -c sleep -- ls -C /jindra/resources/transit)" "one  two"
  ${kubectl} delete job jindra.transit-test.${build_no}
}

it_should_not_merge_two_transit_directories() {
  local build_no=${RANDOM}
  ${jindra_cli} -build-no ${build_no} -c transit-test2.yaml all|${kubectl} apply -f-
  pod=$(wait_for_container transit-test ${build_no} three sleep)

  assert_true "test -n '${pod}'" "container three/sleep should be running"
  assert_eq "$(${kubectl} exec ${pod} -c sleep -- ls -C /jindra/resources/transit)" "two"
  ${kubectl} delete job jindra.transit-test.${build_no}
}

run_tests jindra
