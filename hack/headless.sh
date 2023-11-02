#!/bin/bash

host="${HOST:-localhost:8080}"
pid=$$

output="/tmp/report.yaml"
binary=""
artifact=""
gitURL=""
gitUser=""
gitPassword=""
gitKey=""


appId=""
taskId=""


usage() {
  echo "Usage:"
  echo "  -h help"
  echo "  -b binary path."
  echo "  -a artifact."
  echo "  -o output path."
  echo "  -r Git repository URL."
  echo "  -U Git user."
  echo "  -P Git password."
  echo "  -K Git SSH key path."
}

while getopts "u:b:a:r:U:P:K:h" arg; do
  case $arg in
    h)
      usage
      exit 1
      ;;
    u)
      host=$OPTARG
      ;;
    o)
      output=$OPTARG
      ;;
    b)
      binary=$OPTARG
      ;;
    a)
      artifact=$OPTARG
      ;;
    r)
      gitURL=$OPTARG
      ;;
    U)
      gitUser=$OPTARG
      ;;
    P)
      gitPassword=$OPTARG
      ;;
    K)
      gitKey=$OPTARG
      ;;
  esac
done

echo "==================================="
echo "Hub URL: ${host}"
echo "Output path: ${output}"
echo "Binary path: ${binary}"
echo "Artifact: ${artifact}"
echo "Git URL: ${gitURL}"
echo "Git User: ${gitUser}"
echo "Git Password: ${gitPassword}"
echo "Git SSH Key path: ${gitKey}"
echo "==================================="

set -e

#
# Create application.
#
d="
---
name: Temporary-${pid}
description: Temporary application.
binary: ${artifact}
"
if [ ! -z "$gitURL" ]
then
d+="
repository:
  kind: git
  url: ${gitURL}
"
fi
appId=$(curl -fs -X POST ${host}/applications -H 'Content-Type:application/x-yaml' -d "${d}" | jq .id)
echo "Application: ${appId} created."


#
# Create task.
#
d="
---
addon: analyzer
application:
  id: ${appId}
data:
  artifact: ${artifact}
  binary: $(basename "${binary}")
  rules:
    labels:
      included:
      - konveyor.io/target=cloud-readiness
"
taskId=$(curl -fs -X POST ${host}/tasks -H 'Content-Type:application/x-yaml' -d "${d}" | jq .id)
echo "Task ${taskId} created."
#
# Upload the binary into the task bucket.
#
if [ ! -z "${binary}" ]
then
  curl -fs -F "file=@${binary}" ${host}/tasks/${taskId}/bucket/$(basename "${binary}")
  echo "Binary ${binary} uploaded."
fi

#
# Submit task.
#
curl -f -X PUT ${host}/tasks/${taskId}/submit
echo "Task ${taskId} submitted."


#
# Fetch the task.
#
done=0
lastState=""
while [ ${done} -eq 0 ]
do
  sleep 4
  state=$(curl -fs ${host}/tasks/${taskId} | jq .state | tr -d '"')
  if [ "${state}" != "${lastState}" ]
  then
    lastState=${state}
    echo ${state}
  else
    echo -n "."
  fi
  case ${state} in
    Succeeded)
      done=1
      ;;
    Failed | Canceled)
      exit 1
      ;;
  esac
done

#
# Download report.
#
curl -fs --output ${output} ${host}/applications/${appId}/analysis -H 'Accept:application/x-yaml'
echo "Report downloaed: ${output}"

#
# Delete app.
#
curl -f -X DELETE ${host}/applications/${appId}
echo "Application: ${appId} deleted."




