#!/bin/bash

echo "$(git symbolic-ref -q --short HEAD || git describe --tags --exact-match)"
