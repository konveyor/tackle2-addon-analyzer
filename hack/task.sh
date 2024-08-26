#!/bin/bash

host="${HOST:-localhost:8080}"
state="${1:-Ready}"

curl -X POST ${host}/tasks \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
---
state: ${state}
kind: analyzer
application:
  id: 1
data:
  mode:
    withDeps: true
  rules:
    labels:
      included:
        - konveyor.io/target=cloud-readiness
  tagger:
    enabled: true
"
