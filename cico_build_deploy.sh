#!/bin/bash

/usr/sbin/setenforce 0

# Get all the deps in
yum -y install \
    docker \
    make \
    git \
    curl

sed -i '/OPTIONS=.*/c\OPTIONS="--selinux-enabled --log-driver=journald --insecure-registry registry.ci.centos.org:5000"' /etc/sysconfig/docker
service docker start


make docker-build-build
make docker-install
make docker-test
make docker-build-run
make docker-run-deploy