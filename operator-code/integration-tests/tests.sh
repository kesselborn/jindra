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
 for f in test_*.sh; do include $f; done

# execute commands before running any test
before_all() {
  kubectl create namespace ${NAMESPACE}
  ${kubectl} apply -f ../playground/jindra-runner-permissions.yaml

  local pod=$(wait_for_jindra_operator)
  assert_true "test -n '${pod}'" "operator pod should be running"
}

# execute command when all tests have beend executed
after_all() {
  kubectl delete --wait=false namespace ${NAMESPACE}
}

run_tests jindra
