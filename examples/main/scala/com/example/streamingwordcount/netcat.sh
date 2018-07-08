#!/usr/bin/env bash
[ "$DEBUG" = "1" ] && set -x
set -euo pipefail
err_report() { echo "errexit on line $(caller)" >&2; }
trap err_report ERR
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

APP_LABEL=streamingwordcount

# make sure driver is dead before starting listener
kubectl delete po -lapp=${APP_LABEL},spark-role=driver||true

# tail logs in the background
stern --tail 10 --color always -l app=$APP_LABEL &

echo "Starting netcat on port 9999 and connecting input to stdin..."
nc -lk 9999
