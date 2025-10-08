#!/bin/sh

cd "$(dirname "$0")"
./server -e athens
echo "Press any key to exit..."
read dummy