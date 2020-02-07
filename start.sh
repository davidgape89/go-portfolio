#!/bin/bash

go run main.go &

inotifywait -q -m -e close_write *.go **/*.go |
while read -r file event; do
    if [ "$!" ] 
    then
        kill $!
    fi

    if [ "$(ps | grep main | awk '{print $1}')" ]
    then 
        kill $(ps | grep main | awk '{print $1}')
    fi
    echo "Recompiling..."
    go run main.go &
done