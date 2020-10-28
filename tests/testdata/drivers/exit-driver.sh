#!/bin/bash

# stub out support for imageType of "docker"
if [[ "$@" == "--handles" ]]; then
  echo "docker"
else
  exit 0
fi