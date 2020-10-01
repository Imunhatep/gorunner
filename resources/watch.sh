#!/usr/bin/env bash

set -e

watch_file_changes() {
  FILE="$1"

  echo "Watching: ${FILE}"

  if [ ! -e "${FILE}" ]; then
    echo "Error: file not found"
    exit 2
  fi

  LAST=`md5sum "$FILE"`

  while true; do
    sleep 1
    NEW=`md5sum "${FILE}"`
    if [ "${NEW}" != "${LAST}" ]; then
      echo "${FILE} content change detected, exiting!"
      exit 1
    else
      echo "${FILE} keep.."
    fi
  done
}

FILE=$1
watch_file_changes "${FILE}"
