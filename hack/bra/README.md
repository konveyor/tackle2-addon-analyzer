## Overview ##

Add scripts for support bulk analysis of binaries found in a directory (tree).

Assumptions:
- konveyor auth is disabled.
- binary file names do not contain spaces.
- All files in the directory are binaries.
- 0,1 credentials is required and will be created by the user prior to running scripts.
- The -d _url_ is the route to the konveyor application found with:
  - minikube ip
  - oc get route 
 - User batch binaries as desired using organization of the file tree.
 - Labels (sources and targets):
      - konveyor.io/source=javaee
      - konveyor.io/target=cloud-readiness
      - konveyor.io/target=openjdk17
      - konveyor.io/target=openliberty
      - konveyor.io/target=quarkus

 
Dependencies:
- bash 4+
- jq

## Tools ##

creds.sh - Tool for listing credentials.
---
```
 ./creds.sh -h
Usage: creds.sh
  -u konveyor URL
  -h help
```
Example:
```
$ ./creds.sh -u http://192.168.49.2
ID  | Kind      | Name
--- | ----------|------------------
1     maven       Test
2     source      My
```

**analysis.sh** - Tool for running analysis or reporting status of analysis tasks.
```
$ ./analysis.sh -h
Usage: analysis.sh <required> <action> <options>
  -h help
Required
  -u URL.
  -d directory of binaries.
Actions:
  -s show summary.
  -r run analysis.
  -x cancel tasks.
  -l list applications with status.
Options:
  -c credentials id.
  -f forced. used with -r
  -o output
```
Example: Run analysis on directory with 3 jar files. 
Note: without the `-f` option, an analysis task will only be created when there isn't an existing task in-flight or succeeded.
```
$ ./analysis.sh -d ../jars/small -u http://192.168.49.2 -r
Application for: ../jars/small/cat.jar created.
Application for: ../jars/small/dog.jar created.
Application for: ../jars/small/tiger.jar created.

Task for: ../jars/small/cat.jar created. id=35
Artifact: ../jars/small/cat.jar uploaded. id=35
Task for: ../jars/small/cat.jar submitted. id=35

Task for: ../jars/small/dog.jar created. id=36
Artifact: ../jars/small/dog.jar uploaded. id=36
Task for: ../jars/small/dog.jar submitted. id=36

Task for: ../jars/small/tiger.jar created. id=37
Artifact: ../jars/small/tiger.jar uploaded. id=37
Task for: ../jars/small/tiger.jar submitted. id=37
```
Example: report status.
```
$ ./analysis.sh -d ../jars -u http://192.168.49.2 -ls

    Count: 3
  Created: 0
  Pending: 1
Postponed: 0
  Running: 2
Succeeded: 0
   Failed: 0

ID  | State     | Application
--- | ----------|------------------
1     Running     cat.jar
2     Pending     dog.jar
3     Running     tiger.jar

```

owner.sh - Tool for assigning application owner (stakeholder)
---
The stakeholder is created (as needed).
Processes a directory of text files containing a list of binaries.
The directory part of the application name is ignored when matching.  Example: dog.war matches both a/dog.war b/dog.war.
The file name (suffix ignored) is used as the name of the stakeholder.
```
$ ./owner.sh -h
Usage: owner.sh <required> <options>
  -h help
Required
  -u URL.
  -d directory of binaries.
Actions:
  -a assign owner.
Options:
  -o output
```
Example:
```
$  tree ../streams/
../streams/
├── animals
└── fruit

1 directory, 2 files
$ cat ../streams/animals 
dog.jar
cat.jar
lion.jar
$ cat ../streams/fruit 
apple.jar
lemon.jar
orange.jar
```
```
$ ./owner.sh  -d ../streams -u http://192.168.49.2 -a
stakeholder for: animals created. id=1
stakeholder for: fruit created. id=2
*/dog.jar (id=2) assigned owner animals (id=1)
*/dog.jar (id=8) assigned owner animals (id=1)
*/cat.jar (id=1) assigned owner animals (id=1)
application for: ../streams/animals:3 "lion.jar" - NOT FOUND
*/apple.jar (id=5) assigned owner fruit (id=2)
*/lemon.jar (id=6) assigned owner fruit (id=2)
*/orange.jar (id=7) assigned owner fruit (id=2)
```

tag.sh - Tool to add tags to applications.
---
same input files as `owner.sh`.
```
$ ./tag.sh -h
Usage: tag.sh <required> <options>
  -h help
Required
  -u URL.
  -d directory of binaries.
Actions:
  -c create category and tags assigned to applications.
  -x DELETE category and tags.
Options:
  -o output
```

Example:
```
$ ./tag.sh -d ../streams -u http://192.168.49.2 -c MyStream2
tag category: MyStream created. (id=56)
tag: MyStream2=fruit created. (id=450)
tag MyStream2=fruit (id=450) added to application */apple.jar (id=5)
tag MyStream2=fruit (id=450) added to application */lemon.jar (id=6)
tag MyStream2=fruit (id=450) added to application */orange.jar (id=7)
tag: MyStream2=animals created. (id=451)
tag MyStream2=animals (id=451) added to application */dog.jar (id=2)
tag MyStream2=animals (id=451) added to application */dog.jar (id=8)
tag MyStream2=animals (id=451) added to application */cat.jar (id=1)
tag MyStream2=animals (id=451) added to application */tiger.jar (id=3)
application for: ../streams/animals:4 "lion.jar" - NOT FOUND
```

The tag category can be deleted using **-x** option. This will delete all associated tags (and association with applications).
```
$ ./tag.sh -u http://192.168.49.2 -x MyStream2
tag category: MyStream2 DELETED. (id=53)
```

wave.sh - Tool for creating migration waves and associating applications.
---
same input files as `owner.sh`.
```
$ ./wave.sh -h
Usage: wave.sh <required> <options>
  -h help
Required
  -u URL.
  -d directory of binaries.
Actions:
  -a assign applications to waves.
  -x DELETE waves.
Options:
  -s start date. Eg: 2024-03-13T09:01:24-07:00
  -e end date.   Eg: 2024-03-14T09:01:24-07:00
  -o output
```
Examples:  note: The -s <start> -e <end> will default when not specified.
```
$ ./wave.sh -d ../streams -u http://192.168.49.2 -a
wave: fruit created. id=1
wave: animals created. id=2
wave fruit updated. (id=1) with 3 applications.
application for: ../streams/animals:4 "lion.jar" - NOT FOUND
wave animals updated. (id=2) with 4 applications.
```
```
$ ./wave.sh -d ../streams -u http://192.168.49.2 -x
wave: fruit DELETED. id=1
wave: animals DELETED. id=2
```

report.sh - Tool to download html reports.
---

Processes the tree of binaries used to perform analysis.
Will not overwrite existing reports.  This can be overridden using the `-f forced` option.
The (-r <path) will be the root of the output directory tree which replicates the structure of the input tree (of binaries).

```
$ ./report.sh -h
Usage: report.sh <required> <options>
  -h help
Required
  -u URL.
  -d directory of binaries.
  -r report directory.
  -f forced.
Options:
  -o output
```

Example:
```
$ ./report.sh -d ../jars -u http://192.168.49.2  -r /tmp/jeff
ID  | STATUS     | PATH (destination)
--- | -----------|-----------------------------
1     SUCCEEDED    /tmp/jeff/hack/jars/cat.jar.tgz
2     SUCCEEDED    /tmp/jeff/hack/jars/dog.jar.tgz
3     SUCCEEDED    /tmp/jeff/hack/jars/tiger.jar.tgz
4     SUCCEEDED    /tmp/jeff/hack/jars/apple.jar.tgz
5     SUCCEEDED    /tmp/jeff/hack/jars/lemon.jar.tgz
6     SUCCEEDED    /tmp/jeff/hack/jars/orange.jar.tgz
```

**note**: 
- application of _lion.jar_ does not exist to demonstrate error reporting.
- application dog.jar and ../jars/dog.jar both exist.
