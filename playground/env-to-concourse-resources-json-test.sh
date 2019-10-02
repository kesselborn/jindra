SCRIPT=./env-to-concourse-resources-json.sh

it_should_return_empty_json_when_no_env_matches() {
  res=$(env ${SCRIPT})
  assert_eq "${res}" '{}'
}

it_should_convert_source_keys() {
  res=$(env source_foo=bar source_bar=baz ${SCRIPT})
  assert_eq "${res}" '{"source":{"bar":"baz","foo":"bar"}}'
}

it_should_convert_values_with_whitespace_correctly() {
  res=$(env source_foo="bar 	foo" source_bar=baz ${SCRIPT})
  assert_eq "${res}" '{"source":{"bar":"baz","foo":"bar 	foo"}}'
}


it_should_convert_version_keys() {
  res=$(env version_ref=bar  ${SCRIPT})
  assert_eq "${res}" '{"version":{"ref":"bar"}}'
}

it_should_convert_source_and_version_keys() {
  res="$(env source_uri=bar source_private_key="-----BEGIN OPENSSH PRIVATE KEY-----\\n....AQIDBA==\\n-----END OPENSSH PRIVATE KEY-----" version_ref=asdgasdg  ${SCRIPT})"
  assert_eq "${res}" '{"source":{"private_key":"-----BEGIN OPENSSH PRIVATE KEY-----\n....AQIDBA==\n-----END OPENSSH PRIVATE KEY-----","uri":"bar"},"version":{"ref":"asdgasdg"}}'
}

it_should_convert_source_version_and_params_keys() {
  res=$(env source_uri=bar source_private_key="key" params_depth="5" version_ref=asdgasdg  ${SCRIPT})
  assert_eq "${res}" '{"source":{"private_key":"key","uri":"bar"},"version":{"ref":"asdgasdg"},"params":{"depth":"5"}}'
}

