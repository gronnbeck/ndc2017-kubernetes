#!/bin/bash

kubectl create secret generic ssl-tunnel-priv-key --from-file=$HOME/.ssh/compose
