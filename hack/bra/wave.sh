#!/bin/bash
#
# Assign applications to a migration wave.
# The wave is created (as needed).
# Processes a directory of text files containing a list of binaries.
# The file name (suffix ignored) is used as the name of the wave.
# Example:
# $ ls -Fal streams
# waveA.txt
# waveB.txt
# waveC.txt
# 
# $ cat streams/waveA.txt
# dog.war
# cat.war
# tiger.war
#

pid=$$
self=$(basename $0)
tmp=/tmp/${self}-${pid}

declare -A applications
declare -A waves


usage() {
  echo "Usage: ${self} <required> <options>"
  echo "  -h help"
  echo "Required"
  echo "  -u URL."
  echo "  -d directory of binaries."
  echo "Actions:"
  echo "  -a assign application to wave. (default)"
  echo "Options:"
  echo "  -s start date. Eg: 2024-03-13T09:01:24-07:00"
  echo "  -e end date.   Eg: 2024-03-14T09:01:24-07:00"
  echo "  -o output"
}

while getopts "u:d:s:e:ha" arg; do
  case $arg in
    u)
      host=$OPTARG/hub
      ;;
    d)
      dirPath=$OPTARG
      ;;
    a)
      actionAssign=true
      ;;
    s)
      startDate=$OPTARG
      ;;
    e)
      endDate=$OPTARG
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

if [ -z "${startDate}"  ]
then
  startDate="2025-01-01T09:01:00-00:00"
fi
if [ -z "${endDate}"  ]
then
  endDate="2025-03-01T09:01:00-00:00"
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
          name=$(basename ${name})
          ids=${applications["${name}"]}
          if [ -z "${ids}" ]
          then
            ids=(${id})
          else
            ids+=(${id})
          fi
          applications["${name}"]=${ids[@]}
        fi
      done
      ;;
    *)
      print "find applications - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}

findWaves() {
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/migrationwaves)
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
          waves["${name}"]=${id}
        fi
      done
      ;;
    *)
      print "find waves - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


ensureWaveCreated() {
  name=$1
  if [ -n "${waves[${name}]}" ]
  then
    print "wave: ${name} found."
    return
  fi
  d="
---
name: ${name}
startDate: ${startDate}
endDate: ${endDate}
"
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X POST ${host}/migrationwaves -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    201)
      waveId=$(cat ${tmp}|jq .id)
      waves["${name}"]=${waveId}
      print "wave for: ${name} created. id=${waveId}"
      ;;
    409)
      print "wave for: ${name} found."
      ;;
    *)
      print "create wave - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


ensureWavesCreated() {
  for p in $(find ${dirPath} -type f)
  do
    p=$(basename ${p})
    name="${p%.*}"
    ensureWaveCreated ${name}
  done
}


updateWave() {
  waveName=$1
  waveId=$2
  shift
  shift
  appIds=("$@")
  refs=()
  for id in "${appIds[@]}"
  do
    refs+=("\"id\":${id}")
  done
  refs=($(IFS=, ;echo "${refs[*]}"))
  d="
---
id: ${waveId}
name: ${waveName}
startDate: ${startDate}
endDate: ${endDate}
applications: [${refs[@]}]
"
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X PUT ${host}/migrationwaves/${waveId} -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    204)
      print "wave ${waveName} updated. (id=${waveId})"
      ;;
    *)
      print "assign wave - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac

}


updateWaves() {
  for p in $(find ${dirPath} -type f)
  do
    wave="${p#.*}"
    wave=$(basename ${wave})
    waveId=${waves["${wave}"]}
    n=0
    ids=()
    while read -r entry
    do
      ((n++))
      entry=$(basename ${entry})
      appIds=(${applications[${entry}]})
      if [ -z "${appIds}" ]
      then
        print "application for: ${p}:${n} \"${entry}\" - NOT FOUND"
        continue
      fi
      ids+=(${appIds[@]})
    done < ${p}
    updateWave ${wave} ${waveId} "${ids[@]}"
  done
}


if [ -n "${actionAssign}" ]
then
  findApps
  findWaves
  ensureWavesCreated
  updateWaves
  exit 0
fi

echo -e "\nNo action selected."
usage

