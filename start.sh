#!/bin/bash

go run main.go &

inotifywait -q -m -e close_write *.go **/*.go |
while read -r file event; do
    kill -9 $!
    kill $(ps | grep main | awk '{print $1}')
    echo "Recompiling..."
    go run main.go &
done