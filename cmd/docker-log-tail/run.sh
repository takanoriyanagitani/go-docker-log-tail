#!/bin/sh

export ENV_DO_FOLLOW=true
export ENV_SHOW_STDOUT=true
export ENV_SHOW_STDERR=true
export ENV_SHOW_TIME=true
export ENV_SHOW_DETAILS=true
export ENV_SINCE=
export ENV_UNTIL=
export ENV_TAIL=3

sample_container_name=cadvisor

containerId=${1:-${sample_container_name}}

test 0 -lt ${#containerId} || exec sh -c 'echo container id unknown; exit 1'

./docker-log-tail "${containerId}"
