#!/bin/sh
set -u
mkdir -p ssh-keystore/
ssh-keygen -q -N "" -f ssh-keystore/test
mkdir -p config/.ssh/
cat ssh-keystore/test.pub > config/.ssh/authorized_keys
