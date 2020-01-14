#!/usr/bin/env sh

set -e

if [ -z ${IMAGE_VERSION} ]; then
    echo "IMAGE_VERSION env var needs to be set"
    exit 1
fi

echo ${IMAGE_VERSION}

DIR="$( cd "$( dirname "${0}" )" && pwd )"
ROOT_DIR=${DIR}/..
REPOSITORY="harbor.spotahome.net/devops/"
IMAGE="mandrill-prometheus-exporter"
TARGET_IMAGE=${REPOSITORY}${IMAGE}


docker build \
    --build-arg operator=${OPERATOR} \
    -t ${TARGET_IMAGE}:${IMAGE_VERSION} \
    -f ${ROOT_DIR}/docker/Dockerfile .

if [ -n "${PUSH_IMAGE}" ]; then
    echo "pushing ${TARGET_IMAGE} images..."
    docker push ${TARGET_IMAGE}
fi
