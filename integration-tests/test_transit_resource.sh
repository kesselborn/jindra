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
  defer "${kubectl} delete job jindra.transit-test.${build_no}"
  pod=$(wait_for_container transit-test2 ${build_no} three sleep)

  assert_true "test -n '${pod}'" "container three/sleep should be running"
  assert_eq "$(${kubectl} exec ${pod} -c sleep -- ls -C /jindra/resources/transit)" "two"
}
