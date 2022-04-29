#!/bin/bash

source versions.sh

set -ex

docker build --build-arg C3OS_VERSION=$C3OS_VERSION \
             --build-arg K3S_VERSION=$K3S_VERSION \
             --build-arg LUET_VERSION=$LUET_VERSION \
             --build-arg OS_LABEL=$OS_LABEL \
             --build-arg OS_NAME=$OS_NAME \
             -t $IMAGE \
             -f Dockerfile.${FLAVOR} ./

docker run -v $PWD:/cOS \
           -v /var/run:/var/run \
           -i --rm quay.io/costoolkit/elemental:v0.0.14-e4e39d4 --name $ISO --debug build-iso --date=false --overlay-iso /cOS/overlay/files-iso $IMAGE --output /cOS/