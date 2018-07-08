#!/usr/bin/env bash
[ "$DEBUG" = "1" ] && set -x
set -euo pipefail
err_report() { echo "errexit on line $(caller)" >&2; }
trap err_report ERR

VERSION_TAG=$(date -u +v%Y%m%d)-$(git describe --tags --always --dirty)
K8S_CLUSTER=$(kubectl config current-context)
IMAGE_CHROOT=${IMAGE_CHROOT:="registry.kube-system.svc.cluster.local:80"}

cat <<EOF
STABLE_IMAGE_CHROOT $IMAGE_CHROOT
STABLE_VERSION_TAG $VERSION_TAG
STABLE_K8S_CLUSTER $K8S_CLUSTER
EOF