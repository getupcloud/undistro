#!/usr/bin/env bash

ssh -i $3 ubuntu@$2 -o "ProxyCommand ssh -W %h:%p -i $3 ubuntu@$1"