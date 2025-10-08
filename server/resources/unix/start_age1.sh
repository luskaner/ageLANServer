#!/bin/sh

cd "$(dirname "$0")"
./server -e age1
echo "Press any key to exit..."
read dummy