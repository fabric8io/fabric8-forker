#!/bin/bash

make docker-build-build
make docker-install
make docker-test
make docker-build-run
make docker-run-deploy