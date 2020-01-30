#!/bin/bash

./dummy.sh &

echo $!

echo Kill process?

read var1

while [ "$var1" != "y" ]:
do
    read var1
done

kill $!