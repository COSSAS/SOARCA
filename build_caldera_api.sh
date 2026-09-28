#!/bin/bash
# Before running: 
# - go install github.com/go-swagger/go-swagger/cmd/swagger@latest

set -e
PROJ_PATH="$PWD"


CALDERA_VERSION="e4a4b9a320e707e58e69e3213ca07fdc561350ce" # Commit hash
OUTPUT_BASE_FOLDER="./internal/capability/caldera/api"

OUTPUT="$PROJ_PATH/caldera_api_raw.json"

# DOCKER_CMD="podman"
DOCKER_CMD="docker"

CALDERA_PORT="8888"
CALDERA_IMG_NAME="caldera:latest"
CALDERA_REPO="https://github.com/mitre/caldera"
TMP_WORKDIR="/tmp/caldera_api_build/"

if [ -d "$TMP_WORKDIR" ]; then
    rm -rf $TMP_WORKDIR
fi

mkdir $TMP_WORKDIR
cd $TMP_WORKDIR

git clone $CALDERA_REPO caldera
cd caldera/
git checkout $CALDERA_VERSION
git submodule init
git submodule update --recursive

$DOCKER_CMD build . -t $CALDERA_IMG_NAME
CALDERA_CNTR_ID=$($DOCKER_CMD run -d -p $CALDERA_PORT:8888 $CALDERA_IMG_NAME)

# https://stackoverflow.com/questions/11904772/how-to-create-a-loop-in-bash-that-is-waiting-for-a-webserver-to-respond#21189440
# Interesting note: caldera doesn't like `HEAD`, and using that HTTP verb will cause curl to fail
until curl --output /dev/null --silent --fail http://127.0.0.1:$CALDERA_PORT/api/docs/swagger.json; do
    printf '.'
    sleep 5
done

printf 'CALDERA seems to be up!'

curl http://127.0.0.1:$CALDERA_PORT/api/docs/swagger.json --output $OUTPUT

$DOCKER_CMD stop -t0 $CALDERA_CNTR_ID
$DOCKER_CMD container rm -f $CALDERA_CNTR_ID
$DOCKER_CMD image rm -f $CALDERA_IMG_NAME

cd $PROJ_PATH

rm -rf $TMP_WORKDIR

python3 caldera_swagger_cleanup.py

go run github.com/go-swagger/go-swagger/cmd/swagger@latest generate client -f caldera_api.json -t $OUTPUT_BASE_FOLDER

printf 'Done!'
