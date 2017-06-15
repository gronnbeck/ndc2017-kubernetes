#!/bin/bash 

_project=$1
_cluster=$2

gcloud --project=$_project container clusters get-credentials $_cluster
