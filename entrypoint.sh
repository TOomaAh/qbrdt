#!/bin/sh
if [ ! -f /config/config.yml ]; then
    touch /config/config.yml
    echo "config.yml not found, creating a new one"
    echo "you must edit the file to add your configuration"
    exit 1
fi

# set '/config/config.yml' as CONFIG_FILE env variable
export CONFIG_FILE=/config/config.yml
export QBRDT_DB=/config/qbrdt.db

# Run the command
./qbrdt