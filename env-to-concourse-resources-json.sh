#!/bin/sh
# yeah: this script is horrible, but it needs to function in nearly all
# environments ... so: jq is not a requirement we can work with :/
IFS="
"
env_vars=$(env|sort)

# call: <prefix> <env1> <env2> ...
convert() {
  prefix=$1
  shift
  json=""

  while test -n "$1"
  do
    echo $1|grep "^${prefix}_" >/dev/null|| { shift; continue; }
    kv=$(echo "$1"|sed "s/^${prefix}_//g")
    json="${json}${json:+,}\"$(echo ${kv}|cut -f1 -d=)\":\"$(echo ${kv}|cut -f2-1000 -d=)\""
    shift
  done

  echo "${json}"
}

source=$(convert source ${env_vars})
version=$(convert version ${env_vars})
params=$(convert params ${env_vars})

json="${json}${source:+"\"source\":{${source}}"}"
json="${json}${json:+,}"

json="${json}${version:+"\"version\":{${version}}"}"
json="${json}${json:+,}"

json="${json}${params:+"\"params\":{${params}}"}"
json="${json}${json:+,}"

json=$(echo ${json}|sed 's/,*$//')

echo "{${json}}"
