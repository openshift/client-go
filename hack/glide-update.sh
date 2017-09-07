#!/bin/bash -e

glide update

# need to remove the vendor tree from our types to be able to build against the "right" apimachinery and such for now
glide install --strip-vendor github.com/openshift/origin