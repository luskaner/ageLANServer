#!/bin/sh

cd "$(dirname "$0")"
./server -e age3
read -p "Press any key to exit..."