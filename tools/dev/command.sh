#! /bin/bash

echo "first arg: $1" >&2

for number in {0..5}; do
    echo "$number "
    sleep 1
done

echo "completed"
