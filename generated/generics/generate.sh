#!/bin/bash

if [ -z $GOPATH ]; then
  echo 'GOPATH is not set'
  exit 1
fi

GENERIC_MAP_PATH=${GOPATH}/src/github.com/m3db/m3ninx/vendor/github.com/m3db/m3x/generics/hashmap
GENERIC_MAP_IMPL=${GENERIC_MAP_PATH}/map.go

if [ ! -f "$GENERIC_MAP_IMPL" ]; then
  echo "${GENERIC_MAP_IMPL} does not exist"
  exit 1
fi

GENERATED_PATH=${GOPATH}/src/github.com/m3db/m3ninx/index/segment/mem
if [ ! -d "$GENERATED_PATH" ]; then
  echo "${GENERATED_PATH} does not exist"
  exit 1
fi

mkdir -p $GENERATED_PATH/postingsgen
cat $GENERIC_MAP_IMPL | genny -out=${GENERATED_PATH}/postingsgen/generated_map.go -pkg=postingsgen gen "KeyType=[]byte ValueType=postings.MutableList"

mkdir -p $GENERATED_PATH/fieldsgen
cat $GENERIC_MAP_IMPL | genny -out=${GENERATED_PATH}/fieldsgen/generated_map.go -pkg=fieldsgen gen "KeyType=[]byte ValueType=*postingsgen.ConcurrentMap"