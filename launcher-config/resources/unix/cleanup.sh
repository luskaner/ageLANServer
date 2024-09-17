#!/bin/bash

cd "$(dirname "$0")"
./bin/config revert -a -g
read -p "Press any key to continue..."