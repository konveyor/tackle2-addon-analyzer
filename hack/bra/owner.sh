#!/bin/bash
#
# Assign application owner (stakeholder).
# The stakeholder is created (as needed).
# Processes a directory of text files containing a list of binaries.
# The file name (suffix ignored) is used as the name of the stakeholder.
# Example:
# $ ls -Fal streams
# ownerA.txt
# ownerB.txt
# ownerC.txt
# 
# $ cat streams/ownerA.txt
# dog.war
# cat.war
# tiger.war
#

pid=$$
self=$(basename $0)
tmp=/tmp/${self}-${pid}

declare -A applications
declare -A stakeholders


usage() {
  echo "Usage: ${self} <required> <options>"
  echo "  -h help"
  echo "Required"
  echo "  -u URL."
  echo "  -d directory of binaries."
  echo "Actions:"
  echo "  -a assign owner. (default)"
  echo "Options:"
  echo "  -o output"
}

while getopts "u:d:ha" arg; do
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

findOwners() {
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/stakeholders)
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
          stakeholders["${name}"]=${id}
        fi
      done
      ;;
    *)
      print "find stakeholders - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


ensureOwnerCreated() {
  name=$1
  if [ -n "${stakeholders[${name}]}" ]
  then
    print "stakeholder for: ${name} found."
    return
  fi
  d="
---
name: ${name}
email: "${name}@redhat.com"
"
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X POST ${host}/stakeholders -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    201)
      ownerId=$(jq .id ${tmp})
      stakeholders["${name}"]=${ownerId}
      print "stakeholder for: ${name} created. id=${ownerId}"
      ;;
    409)
      print "stakeholder for: ${name} found."
      ;;
    *)
      print "create skakeholder - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


ensureOwnersCreated() {
  for p in $(find ${dirPath} -type f)
  do
    p=$(basename ${p})
    name="${p%.*}"
    ensureOwnerCreated ${name}
  done
}

assignOwner() {
  ownerName=$1
  ownerId=$2
  appName=$3
  appId=$4
  d="
---
owner:
  id: ${ownerId}
" 
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X PUT ${host}/applications/${appId}/stakeholders -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    204)
      print "${appName} (id=${appId}) assigned owner ${ownerName} (id=${ownerId})"
      ;;
    *)
      print "assign owner - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac

}

assignOwners() {
  for p in $(find ${dirPath} -type f)
  do
    owner="${p#.*}"
    owner=$(basename ${owner})
    ownerId=${stakeholders["${owner}"]}
    n=0
    while read -r entry
    do
      ((n++))
      entry=$(basename ${entry})
      appIds=(${applications[${entry}]})
      if [ "${#appIds[@]}" -eq 0 ]
      then
        print "application for: ${p}:${n} \"${entry}\" - NOT FOUND"
        continue
      fi
      for appId in "${appIds[@]}"
      do
        assignOwner ${owner} ${ownerId} "*/${entry}" ${appId}
      done
    done < ${p}
  done
}


if [ -n "${actionAssign}" ]
then
  findApps
  findOwners
  ensureOwnersCreated
  assignOwners
  exit 0
fi

echo -e "\nNo action selected."
usage

