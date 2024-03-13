#!/bin/bash
# Add tag to application.
# Both the category and tag are created (as needed).
# Processes a directory of text files containing a list of binaries.
# The file name (suffix ignored) is used as the name of the tag.
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
declare -A categories
declare -A tags


usage() {
  echo "Usage: ${self} <required> <options>"
  echo "  -h help"
  echo "Required"
  echo "  -u URL."
  echo "  -d directory of binaries."
  echo "  -c category."
  echo "  -x DELETE tag category."
  echo "Options:"
  echo "  -o output"
}

while getopts "u:d:c:xh" arg; do
  case $arg in
    u)
      host=$OPTARG/hub
      ;;
    d)
      dirPath=$OPTARG
      ;;
    c)
      category=$OPTARG
      ;;
    x)
      delete=true
      ;;
    h)
      usage
      exit 1
  esac
done

if [ -z "${delete}" ]
then
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
fi

if [ -z "${host}"  ]
then
  echo "-u required."
  usage
  exit 0
fi

if [ -z "${category}"  ]
then
  echo "-c required."
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


findCategories() {
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/tagcategories)
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
        categories["${name}"]=${id}
      done
      ;;
    *)
      print "find categories - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


findTags() {
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/tags)
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    200)
      readarray report <<< $(jq -c '.[]|"\(.id) \(.name) \(.category.name)"' ${tmp})
      for r in "${report[@]}"
      do
        r=${r//\"/}
        a=($r)
        id=${a[0]}
        name=${a[1]}
	cat=${a[2]}
	key="${cat}=${name}"
        tags["${key}"]=${id}
      done
      ;;
    *)
      print "find tags - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


ensureCategory() {
  name=${category}
  catId=${categories["${name}"]}
  if [ -n "${catId}"  ]
  then
    print "tag category: ${name} found."
    return
  fi
  d="
---
name: ${name}
"
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X POST ${host}/tagcategories -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    201)
      catId=$(jq .id ${tmp})
      categories["${name}"]=${catId}
      print "tag category: ${name} created. (id=${catId})"
      ;;
    409)
      print "tag category: ${name} found."
      ;;
    *)
      print "create category - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


deleteCategory() {
  catId=${categories["${category}"]}
  if [ -z "${catId}" ]
  then
    print "category ${category} not-found."
    return
  fi
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X DELETE ${host}/tagcategories/${catId})
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    204)
      print "tag category: ${category} DELETED. (id=${catId})"
      ;;
    *)
      print "DELETE category (id=${catId}) - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


ensureTag() {
  name=$1
  key="${category}=${name}"
  tagId=${tags["${key}"]}
  if [ -n "${tagId}"  ]
  then
    print "tag: ${key} found."
    return
  fi
  d="
---
name: ${name}
category:
 id: ${catId}
"
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X POST ${host}/tags -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    201)
      tagId=$(jq .id ${tmp})
      tags["${key}"]=${tagId}
      print "tag: ${key} created. (id=${tagId})"
      ;;
    409)
      print "tag: ${key} found."
      ;;
    *)
      print "create tag - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


addTag() {
  tag=$1
  tagId=$2
  appName=$3
  appId=$4
  d="
---
id: ${tagId}
"
  code=$(curl -kSs -o ${tmp} -w "%{http_code}" -X POST ${host}/applications/${appId}/tags -H 'Content-Type:application/x-yaml' -d "${d}")
  if [ ! $? -eq 0 ]
  then
    exit $?
  fi
  case ${code} in
    201)
      print "tag ${category}=${tag} (id=${tagId}) added to application ${appName} (id=${appId})"
      ;;
    409)
      print "tag ${category}=${tag} (id=${tagId}) found on application ${appName} (id=${appId})"
      ;;
    *)
      print "assign tag - FAILED: ${code}."
      cat ${tmp}
      exit 1
  esac
}


addTags() {
  for p in $(find ${dirPath} -type f)
  do
    tag="${p#.*}"
    tag=$(basename ${tag})
    ensureTag ${tag}
    key="${category}=${tag}"
    tagId=${tags["${key}"]}
    n=0
    while read -r entry
    do
      ((n++))
      entry=$(basename ${entry})
      appIds=${applications[${entry}]}
      if [ -z "${appIds}" ]
      then
        print "application for: ${p}:${n} \"${entry}\" - NOT FOUND"
        continue
      fi
      for appId in ${appIds[@]}
      do
        addTag ${tag} ${tagId} "*/${entry}" ${appId}
      done
    done < ${p}
  done
}

findCategories
if [ -n "${delete}" ]
then
  deleteCategory
  exit
fi
ensureCategory
findTags
findApps
addTags


