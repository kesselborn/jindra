#!/bin/sh

source spec.sh

: ${KUBECONFIG:=${PWD}/kind.kubeconfig}
: ${TRIES:=100}
: ${NAMESPACE:=jindra-tests-${RANDOM}}

kubectl="kubectl --kubeconfig ${KUBECONFIG} --namespace ${NAMESPACE}"
jindra_cli=../../bin/jindra-cli

function run() {
  local pipeline=$1
  ${jindra_cli} -build-no 42 -c ${pipeline} all|${kubectl} apply -f-
}

# you can put tests in other files, just include them
for f in test_*.sh; do include $f; done

# execute commands before running any test
before_all() {
  assert_true "make -C ../.. bin/jindra-cli"
  assert_true "test -x ${jindra_cli}"
  local pod=$(wait_for_jindra_operator)
  assert_true "test -n '${pod}'" "operator pod should be running"
}

# execute command when all tests have beend executed
after_all() {
  make -C ../.. reset-config
  kubectl delete --wait=false namespace ${NAMESPACE} &>/dev/null
}

run_tests jindra
