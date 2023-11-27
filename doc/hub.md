# Launching an Analysis task from the Hub API

To launch an application analysis from the API, it is necessary to POST a Task resource to the Hub's
`/tasks` API endpoint. What follows is an example Task for running a Windup source analysis.

```json
{
  "name": "my-task-name",
  "addon": "windup",
  "application": {
    "id": 1
  },
  "state": "Ready",
  "data": {
    "mode": {
      "binary": false,
      "diva": false,
      "withDeps": false,
      "artifact": "",
      "csv": true
    },
    "output": "/windup/report",
    "rules": {
      "bundles": [
        {
          "id": 3
        }
      ],
      "repository": {
        "kind": "git",
        "url": "https://my-git-repository",
        "branch": "main",
        "path": "/path/to/rules"
      },
      "identity": {
        "id": 1
      },
      "tags": {
        "excluded": []
      }
    },
    "scope": {
      "packages": {
        "included": [],
        "excluded": []
      },
      "withKnown": false
    },
    "sources": [],
    "targets": ["quarkus"],
    "tagger": {
      "enabled": true
    }
  }
}
```

The above JSON creates a Source Analysis task for Application 1 that uses both a builtin ruleset and a ruleset that is defined in a remote repository, using Quarkus as the target technology. The report will be saved to /windup/report in the Application's bucket and will be available via the bucket endpoint at: `/applications/1/bucket/windup/report`.

## Task resource definition

* `name`: A name for the task
* `addon`: The name of the addon to run. This must match the name of the analysis addon CR. In this case it should be `windup`.
* `application`: This is an object with an `id` key where the value is the ID of the Application to analyze.
* `state`: This must be set to `Created` or `Ready` when creating the Task. The task will start as soon as possible if this is set to `Ready`. If the task is created with the state set to `Created`, a follow up `PUT` to `/tasks/:id/` will be necessary to change the state to `Ready` to start the task. Once the task is started, the Hub will automatically update the state to `Running`, `Succeeded` or `Failed` as appropriate.
* `data`: A structure containing parameters for the analysis.
    - `output`: The path in the application bucket where the analysis report should be generated. This should be `/windup/report` to be consistent with analysis run from the UI. This directory is cleared when starting a new analysis.
    - `mode`:
        - `binary`: Boolean. If true this is a binary analysis, else a source analysis.
        - `artifact`: Path in the application bucket to a previously uploaded binary artifact to analyze instead of the one specified on the Application object. Binary analysis only.
        - `withDeps`: Boolean. analyze dependencies. Source analysis only.
        - `diva`: Boolean. Enables transaction analysis
        - `csv`: Boolean. Generate the report data in CSV format, in addition to HTML.
    - `rules`:
        - `bundles`: list of references to RuleBundles to use (the list of built in rulebundles can be queried from the Hub API's /rulebundles/ endpoint)
        - `repository`: optional repository containing rulesets
            - `branch`
            - `kind`
            - `path`
            - `url`
        - `identity`: Optional. This is an object with an `id` key where the value is the ID of the Identity to use to access the ruleset repository.
        - `tags`:
            - `excluded`: Optional. List of rules tags to exclude.
    - `tagger`:
        - `enabled` - Boolean. Enable automated tagging of the Application based on analysis results.
    - `scope`: - Controls the scope of dependencies to include in the analysis.
        - `packages`:
            - `excluded`: List of packages to exclude
            - `included`: List of packages to include. If this is empty, every package in the application is scanned.
        - `withKnown` - Boolean. Analyze known libraries embedded in your application. By default only application code is analyzed.
    - `sources`: List of source technologies to migrate from. In conjunction with the targets this helps to determine what rulesets are used. The list of builtin sources can be found in the UI.
    - `targets`: List of target technologies to migrate to. In conjunction with the sources this helps to determine what rulesets are used. The list of builtin targets can be found in the UI.

## Task status

The task has a few read-only fields which are updated by the system with the status of the analysis task. The most relevant fields are:

* `pod`: The namespaced name of the pod the task is running in.
* `started`: The time the pod started.
* `terminated`: The time the pod terminated.
* `report`:
    - `status`: One of `Running`, `Succeeded`, or `Failed`.
    - `error`: If the task failed, this will contain the error message.
    - `activity`: A list of activity messages published by the task as it runs.

```json
{
  "pod": "konveyor-tackle/task-11-rbmjc",
  "started": "2023-04-13T12:33:47.946435185Z",
  "terminated": null,
  "report": {
    "status": "Running",
    "error": "",
    "activity": ["Fetching application."]
  }
}
```

## Example cURL command

```shell
curl -X POST ${host}/tasks -d \
'{
    "name":"Windup",
    "state": "Ready",
    "locator": "windup",
    "addon": "windup",
    "application": {"id": 1},
    "data": {
        "mode": {
            "artifact": "",
            "binary": false,
            "withDeps": false,
	    "diva": true
        },
        "output": "/windup/report",
        "rules": {
            "path": "",
            "tags": {
                "excluded": [ ]
            }
        },
        "scope": {
            "packages": {
                "excluded": [ ],
                "included": [ ]
            },
            "withKnown": false
        },
        "sources": [ ],
        "targets": [
            "cloud-readiness"
        ]
    }
}'
```

## Retrieving the Analysis Report

The analysis report will be placed into the application's bucket in the directory path specified by the `output` field in the Task data. The report can be downloaded as a gzipped tarball by GETting the directory from the bucket with an `Accept` header that does not explicitly request `text/html`.

```shell
curl ${host}/applications/1/bucket/windup/report -o report.tar.gz
```

CSV reports can be downloaded separately by using the `filter` query parameter:

```shell
curl "${host}/applications/1/bucket/windup/report?filter=*.csv" -o csv-report.tar.gz
```