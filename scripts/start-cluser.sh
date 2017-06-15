#!/bin/bash
_num_nodes=${2:-18}
gcloud container clusters create $1 \
	--machine-type=n1-standard-16 \
	--num-nodes=$_num_nodes
