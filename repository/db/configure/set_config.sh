#!/bin/bash

cp /etc/postgresql/16/main/postgresql.conf /etc/postgresql/16/main/postgresql.conf.backup

cp config.conf /etc/postgresql/16/main/postgresql.conf

sudo systemctl restart postgresql 
