#!/bin/bash

host="${HOST:-localhost:8080}"
state="${1:-Ready}"

curl -X POST ${host}/taskgroups \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
---
state: ${state}
kind: analyzer
tasks:
- application:
    id: 1
data:
  mode:
    artifact: ""
    binary: false
    withDeps: false
  rules:
    labels:
      excluded: []
      included:
        - konveyor.io/target=cloud-readiness
    path: ""
    tags:
      excluded: []
  scope:
    packages:
      excluded: []
      included: []
    withKnownLibs: false
  sources: []
  tagger:
    enabled: true
  targets: []
  verbosity: 0
"
