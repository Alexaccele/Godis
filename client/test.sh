#!/bin/bash
go build main.go
#for i in {0..2} ; do
#    echo "The key is: $i,value is: $i"
#    ./main.exe -op setT -k $i -v $i -t 30
#done

for i in {6..7} ; do
    echo "The key is: $i,value is: $i"
    ./main.exe -op set -k $i -v $i
done