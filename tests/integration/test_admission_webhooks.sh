# this should be triggered by jindras internal validation
it_should_reject_invalid_pipeline() {
  local res=$(${kubectl} apply --wait -f internal-validation-test.yaml 2>&1)
  defer "${kubectl} delete pipeline internal-validation-test"

  assert_eq "${res}" \
    "Error from server (input resource 'git' referenced in stage 'one' does not exist): error when creating \"internal-validation-test.yaml\": admission webhook \"validating.jindra.io\" denied the request: input resource 'git' referenced in stage 'one' does not exist"
}

# this should be triggered by standard kubernetes pod validation
it_should_reject_invalid_stage_pod() {
  SKIP_TEST
}
it_should_reject_invalid_on_success_pod() {
  SKIP_TEST
}

it_should_set_default_values() {
  assert_true "${kubectl} apply --wait -f internal-modifications-test.yaml"
  defer "${kubectl} delete pipeline internal-modification-test"

  local src=$(${kubectl} get jindrapipeline internal-modifications-test -oyaml)
  assert_match "${src}" "restartPolicy: Never"
  assert_match "${src}" 'jindra.io/build-no-offset: "0"'
}

it_should_update_pipeline() {
  local res=$(${kubectl} apply --wait -f minimal-pipeline.yaml 2>&1)
  defer "${kubectl} delete pipeline minimal-pipeline"

  cat minimal-pipeline.yaml |sed 's/one/two/g'|${kubectl} apply --wait -f-
  assert_eq "$?" "0"

  local src=$(${kubectl} get jindrapipeline minimal-pipeline -oyaml)
  assert_true "${src}" "grep two"
}
