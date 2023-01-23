#!/bin/bash

display_usage(){
    echo -e "\nUsage: $0 [Docker image] [absolute path to folder with files] [absolute path to log folder]"
    echo -e "Example: $0 word-sequences:v0.0.1 /tmp/files /tmp/logs\n"
}

if [  $# -le 2 ] 
then 
    display_usage
	exit 1
fi 

DOCKER_IMAGE=$1
DIR=$2
LOG_DIR=$3

mkdir -p $LOG_DIR

for file in $DIR/*
do
    docker run --read-only -v $DIR:/tmp $DOCKER_IMAGE /tmp/$(basename $file) > $LOG_DIR/$(basename $file).log &
done

wait 
echo "All done"