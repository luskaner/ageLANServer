#!/bin/sh

cd "$(dirname "$0")"
./server -e age2
echo "Press any key to exit..."
read dummy