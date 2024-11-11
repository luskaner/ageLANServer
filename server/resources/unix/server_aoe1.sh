#!/bin/sh

cd "$(dirname "$0")"
./server -e age1
read -p "Press any key to exit..."