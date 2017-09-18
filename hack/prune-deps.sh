#!/bin/bash -e

dep init
dep prune

# we shouldn't have modified anything
git diff-index --name-only --diff-filter=M HEAD | xargs -r git checkout -f
# we need to preserve code-generator and friends
git diff-index --name-only HEAD | grep -F -e 'k8s.io/code-generator' -e 'k8s.io/gengo' | xargs -r git checkout -f

# now cleanup what's dangling
git clean -x  -f -d
