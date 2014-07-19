#!/bin/bash

cp rethinkdb.conf /etc/init
cp vertigo.conf /etc/init

sudo apt-get update -y
sudo add-apt-repository ppa:rethinkdb/ppa -y
sudo apt-get update -y
sudo apt-get install rethinkdb -y

start rethinkdb
sleep 1
start vertigo
