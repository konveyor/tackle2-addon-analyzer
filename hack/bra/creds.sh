#!/bin/bash

pid=$$
self=$(basename $0)
tmp=/tmp/${self}-${pid}

usage() {
  echo "Usage: ${self}"
  echo "  -u konveyor URL"
  echo "  -h help"
}

while getopts "u:h" arg; do
  case $arg in
    h)
      usage
      exit 1
      ;;
    u)
      host=$OPTARG/hub
      ;;
  esac
done

if [ -z "${host}"  ]
then
  echo "-u required."
  usage
  exit 0
fi

code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/identities)
if [ ! $? -eq 0 ]
then
  exit $?
fi
case ${code} in
  200)
    echo "ID  | Kind      | Name"
    echo "--- | ----------|------------------"
    readarray report <<< $(jq -c '.[]|"\(.id) \(.kind) \(.name)"' ${tmp})
    for r in "${report[@]}"
    do
      r=${r//\"/}
      t=($r)
      id=${t[0]}
      kind=${t[1]}
      name=${t[2]}
      printf "%-6s%-12s%s\n" ${id} ${kind} ${name}
    done
    ;;
  *)
    echo "FAILED: ${code}."
    cat ${tmp}
    exit 1
esac

