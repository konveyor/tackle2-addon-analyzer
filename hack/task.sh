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
addon: analyzer
application:
  id: 1
data:
  tagger:
    enabled: "true"
  rules:
    labels:
      included:
      - konveyor.io/target=cloud-readiness
"
