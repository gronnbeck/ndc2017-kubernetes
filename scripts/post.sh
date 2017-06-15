#!/bin/bash

curl -XPOST http://localhost:8080 -d "{\"value\": $1}"
