#!/bin/bash

pid=$$
self=$(basename $0)
tmp=/tmp/${self}-${pid}

declare -A applications


usage() {
  echo "Usage: ${self}"
  echo "  -u URL"
  echo "  -d directory of binaries"
  echo "  -c credentials id"
  echo "  -r report status"
  echo "  -l report status with listing"
  echo "  -C cancel"
  echo "  -f forced"
  echo "  -o output"
  echo "  -h help"
}

while getopts "u:d:c:o:hlfrC" arg; do
  case $arg in
    h)
      usage
      exit 1
      ;;
    d)
      dirPath=$OPTARG
      ;;
    c)
      credId=$OPTARG
      ;;
    u)
      host=$OPTARG/hub
      ;;
    r)
      report=1
      ;;
    f)
      forced=1
      ;;
    C)
      cancelled=1
      ;;
    l)
      reportList=1
      ;;
    o)
      output=$OPTARG
      echo $0 > ${output}
      ;;
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


ensureAppCreated() {
  path=$1
  name=${path}
  if [ -n "${credId}" ]
  then
    cred="
identities:
  - id: ${credId}
"
  fi
  d="
---
name: ${name}
${cred}
"
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X POST ${host}/applications -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    201)
      appId=$(cat ${tmp}|jq .id)
      print "Application for: ${path} created. id=${appId}"
      ;;
    409)
      print "Application for: ${path} found."
      ;;
    *)
      print "create application - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


ensureAppsCreated() {
  for p in $(find ${dirPath} -type f)
  do
    ensureAppCreated ${p}
  done
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


uploadArtifact() {
  taskId=$1
  path=$2
  artifact=$(basename ${path})
  code=$(curl -kSs -w "%{http_code}" -F "file=@${path}" ${host}/tasks/${taskId}/bucket/${artifact})
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    204)
      print "Artifact: ${path} uploaded. id=${taskId}"
      ;;
    *)
      print "artifact upload - FAILED: ${code}."
      exit 1
  esac
}


submitTask() {
  taskId=$1
  path=$2
  code=$(curl -kSs -w "%{http_code}" -X PUT ${host}/tasks/${taskId}/submit)
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    204)
      print "Task for: ${path} submitted. id=${taskId}"
      ;;
    *)
      print "task submit - FAILED: ${code}."
      exit 1
  esac
}


createTask() {
  appId=$1
  path=$2
  artifact=$(basename ${path})
  d="
---
name: ${name}
locator: ${self}
addon: analyzer
application:
  id: ${appId}
data:
  mode:
    binary: true
    artifact: ${artifact}
  rules:
    labels:
      included:
      - konveyor.io/source=javaee
      - konveyor.io/target=cloud-readiness
      - konveyor.io/target=openjdk17
      - konveyor.io/target=openliberty
      - konveyor.io/target=quarkus
"
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X POST ${host}/tasks -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    201)
      taskId=$(cat ${tmp}|jq .id)
      print "\nTask for: ${path} created. id=${taskId}"
      uploadArtifact ${taskId} ${path}
      submitTask ${taskId} ${path}
      ;;
    *)
      print "task create - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


analyzeApps() {
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/tasks)
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    200)
     ;;
   *)
     print "get fetch - FAILED: ${code}."
     cat ${tmp}
     exit 1
  esac
  declare -A task
  readarray report <<< $(jq -c '.[]|"\(.id) \(.state) \(.name)"' ${tmp})
  for r in "${report[@]}"
  do
    r=${r//\"/}
    t=($r)
    name=${t[2]}
    if [ -n "${name}" ]
    then
      task[${name}]="${r}"
    fi
  done
  for p in $(find ${dirPath} -type f)
  do
    name=${p}
    r=${task[${name}]}
    t=($r)
    appId=${applications[${name}]}
    state=${t[1]}
    case ${state} in
      "Pending"|"Postponed"|"Running"|"Succeeded")
        if [ -n "${forced}" ]
        then
          createTask ${appId} ${p}
        fi
        ;;
      *)
        createTask ${appId} ${p}
        ;;
    esac
  done
}

cancelTask() {
  id=$1
  name=$2
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X PUT ${host}/tasks/${id}/cancel)
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    204)
     print "Task ${id} for: ${name} - CANCELED"
     ;;
   *)
     print "get fetch - FAILED: ${code}."
     cat ${tmp}
     exit 1
  esac
}


cancelTasks() {
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/tasks)
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    200)
     ;;
   *)
     print "get fetch - FAILED: ${code}."
     cat ${tmp}
     exit 1
  esac
  declare -A task
  readarray report <<< $(jq -c '.[]|"\(.id) \(.state) \(.name)"' ${tmp})
  for r in "${report[@]}"
  do
    r=${r//\"/}
    t=($r)
    name=${t[2]}
    if [ -n "${name}" ]
    then
      task[${name}]="${r}"
    fi
  done
  for p in $(find ${dirPath} -type f)
  do
    name=${p}
    r=${task[${name}]}
    t=($r)
    appId=${applications[${name}]}
    id=${t[0]}
    state=${t[1]}
    case ${state} in
      "Created"|"Pending"|"Postponed"|"Running")
	cancelTask ${id} ${name}
        ;;
      *)
        ;;
    esac
  done
}


report() {
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/tasks)
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    200)
      ;;
    *)
      print "fetch tasks - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
  declare -A task
  count=0
  created=0
  pending=0
  postponed=0
  running=0
  succeeded=0
  canceled=0
  failed=0
  readarray report <<< $(jq -c '.[]|"\(.id) \(.state) \(.name)"' ${tmp})
  for r in "${report[@]}"
  do
    r=${r//\"/}
    t=($r)
    name=${t[2]}
    if [ -n "${name}" ]
    then
      task[${name}]="${r}"
    fi
  done
  apps=()
  for p in $(find ${dirPath} -type f)
  do
    name=${p}
    apps+=(${name})
  done
  for name in "${apps[@]}"
  do
    ((count++))
    r="${task[${name}]}"
    t=($r)
    state=${t[1]}
    case $state in
      "Created")
        ((created++))
        ;;
      "Pending")
        ((pending++))
        ;;
      "Postponed")
        ((postponed++))
        ;;
      "Running")
        ((running++))
        ;;
      "Succeeded")
        ((succeeded++))
        ;;
      "Failed")
        ((failed++))
        ;;
      "Canceled")
        ((canceled++))
        ;;
    esac
    done
    print ""
    print "    Count: ${count}"
    print "  Created: ${created}"
    print "  Pending: ${pending}"
    print "Postponed: ${postponed}"
    print "  Running: ${running}"
    print "Succeeded: ${succeeded}"
    print " Canceled: ${canceled}"
    print "   Failed: ${failed}"
    if [ -z "${reportList}" ]
    then
      return
    fi
    print ""
    print "ID  | State     | Application"
    print "--- | ----------|------------------"
    for key in "${apps[@]}"
    do
      id="--"
      state="---"
      name=${key}
      r="${task[${key}]}"
      t=($r)
    if [ ${#t[@]} -eq 3 ]
    then
      id=${t[0]}
      state=${t[1]}
      name=${t[2]}
    fi
    status="$(printf "%-6s%-12s%s\n" ${id} ${state} ${name})"
    print "${status}"
  done
}


main() {
  if [ -n "${cancelled}"  ]
  then
    cancelTasks
    return
  fi
  if [ -n "${report}"  ]
  then
    report
    return
  fi
  ensureAppsCreated
  findApps
  analyzeApps
}

main

