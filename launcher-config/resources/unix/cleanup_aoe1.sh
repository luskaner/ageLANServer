#!/bin/bash

cd "$(dirname "$0")"
./bin/config revert -e age1 -a -g
read -p "Press any key to exit..."