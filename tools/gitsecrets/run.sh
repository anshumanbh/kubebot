#!/bin/bash

mkdir /data/repos
cd /data/repos

git clone $1
reponame=`echo $1 | cut -d '/' -f 5 | cut -d '.' -f 1`

cd $reponame

git secrets --install
git secrets --register-aws
git secrets --add 'xoxp-.*'
git secrets --add 'xoxb-.*'
git secrets --scan -r . > $2

exit 0