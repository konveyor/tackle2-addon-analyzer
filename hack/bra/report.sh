#!/bin/bash

pid=$$
self=$(basename $0)
tmp=/tmp/${self}-${pid}

declare -A applications

usage() {
  echo "Usage: ${self} <required> <options>"
  echo "  -h help"
  echo "Required"
  echo "  -u URL."
  echo "  -d directory of binaries."
  echo "  -r report directory."
  echo "  -f forced."
  echo "Options:"
  echo "  -o output"
}

while getopts "u:d:o:r:hf" arg; do
  case $arg in
    u)
      host=$OPTARG/hub
      ;;
    d)
      dirPath=$OPTARG
      ;;
    r)
      reportPath=$OPTARG
      ;;
    f)
      forced=true
      ;;
    o)
      output=$OPTARG
      echo $0 > ${output}
      ;;
    h)
      usage
      exit 1
  esac
done

if [ -z "${dirPath}"  ]
then
  echo "-d required."
  usage
  exit 1
fi
if ! test -d "${dirPath}"
then
  echo "${dirPath} not a directory." 
  exit 1
fi

if [ -z "${host}"  ]
then
  echo "-u required."
  usage
  exit 0
fi

if [ -z "${reportPath}"  ]
then
  echo "-r required."
  usage
  exit 1
fi



print() {
  if [ -n "${output}"  ]
  then
    echo -e "$@" >> ${output}
  else
    echo -e "$@"
  fi
}


findApps() {
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/applications)
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    200)
      readarray report <<< $(jq -c '.[]|"\(.id) \(.name)"' ${tmp})
      for r in "${report[@]}"
      do
        r=${r//\"/}
        a=($r)
        id=${a[0]}
        name=${a[1]}
        if [ -n "${name}" ]
        then
          applications["${name}"]=${id}
        fi
      done
      ;;
    *)
      print "find applications - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


getReport() {
  appId=$1
  binPath=$2
  #
  path=()
  part=(${binPath//\// })
  for p in ${part[@]}
  do
    path+=($(basename $(realpath ${p})))
  done
  path=(${reportPath} "${path[@]}")
  path=$(IFS=/ ; echo "${path[*]}")
  path+=".tar.gz"
  if test -f "${path}" && [ -z "${forced}" ]
  then
    printf "%-6s%-13s%s%s\n" ${appId} "FOUND" ${path}
    return
  fi
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/applications/${appId}/analysis/report)
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    200)
      action=SUCCEEDED
      mkdir -p $(dirname ${path})
      if [ ! $? -eq 0 ]
      then
        exit $?
      fi
      cp ${tmp} ${path}
      if [ ! $? -eq 0 ]
      then
        exit $?
      fi
      ;;
    *)
    action="FAILED:${code}"
  esac
  printf "%-6s%-13s%s%s\n" ${appId} ${action} ${path}

}


getReports() {
  print "ID  | STATUS     | PATH (destination)"
  print "--- | -----------|-----------------------------"
  for p in $(find ${dirPath} -type f)
  do
    name=${p}
    appId=(${applications[${name}]})
    if [ -z "${appId}" ]
    then
      print "application for: ${p}:${n} \"${entry}\" - NOT FOUND"
      continue
    fi
    getReport ${appId} ${name}
  done
}


findApps
getReports

