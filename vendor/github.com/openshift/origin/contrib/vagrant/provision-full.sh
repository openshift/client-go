#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
USERNAME="${1:-vagrant}"

yum update -y
yum install -y docker-io git vim golang e2fsprogs tmux httpie ctags hg bind-utils which

if [[ ! -d /data/src/github.com/openshift/origin ]]; then
  mkdir -p /data/src/github.com/openshift/origin
  chown $USERNAME:$USERNAME /data/src/github.com/openshift/origin
else
  # patch incompatible with fail-over DNS setup
  SCRIPT='/etc/NetworkManager/dispatcher.d/fix-slow-dns'
  if [[ -f "${SCRIPT}" ]]; then
      echo "Removing ${SCRIPT}..."
      rm "${SCRIPT}"
      sed -i -e '/^options.*$/d' /etc/resolv.conf
  fi
  unset SCRIPT
fi

function set_env {
  USER_DIR=$1
  if [[ $(grep GOPATH $USER_DIR/.bash_profile) = "" ]]; then
    touch $USER_DIR/.bash_profile
    echo "export GOPATH=/data" >> $USER_DIR/.bash_profile
    echo "export PATH=\$GOPATH/src/github.com/openshift/origin/_output/local/bin/linux/amd64:\$GOPATH/bin:\$PATH" >> $USER_DIR/.bash_profile
    echo "cd \$GOPATH/src/github.com/openshift/origin" >> $USER_DIR/.bash_profile

    echo "bind '\"\e[A\":history-search-backward'" >> $USER_DIR/.bashrc
    echo "bind '\"\e[B\":history-search-forward'" >> $USER_DIR/.bashrc
  else
    echo "path variables for $USER_DIR already configured"
  fi
}

set_env /home/$USERNAME
set_env /root

systemctl enable docker
systemctl start docker

echo To install etcd, run hack/install-etcd.sh

sed -i s/^Defaults.*requiretty/\#Defaults\ requiretty/g /etc/sudoers
