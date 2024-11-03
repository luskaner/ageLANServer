#!/bin/bash

cd "$(dirname "$0")"
./bin/config revert -e age3 -a -g
read -p "Press any key to exit..."