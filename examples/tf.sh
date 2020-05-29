#!/usr/bin/env bash
set -eu
workspace=$(cd "$(dirname "$0")" && pwd -P)
{
    cd $workspace
    TF_INPUT=0
    TF_IN_AUTOMATION=0
    action="$1"
    case $action in
    "init")
        terraform init -verify-plugins=true -reconfigure
        ;;
    "plan")
        terraform plan -out=tf.plan
        ;;
    "apply")
        terraform apply tf.plan
        ;;
    "lint")
        tflint . --deep
        ;;
    "fmt")
        terraform fmt -recursive
        ;;
    esac
}
