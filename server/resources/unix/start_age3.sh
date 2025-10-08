#!/bin/sh

cd "$(dirname "$0")"
./server -e age3
echo "Press any key to exit..."
read dummy