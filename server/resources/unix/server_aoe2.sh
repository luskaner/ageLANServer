#!/bin/sh

cd "$(dirname "$0")"
./server -e age2
read -p "Press any key to exit..."