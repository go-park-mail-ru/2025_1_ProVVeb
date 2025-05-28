#!/bin/bash

sudo cp /etc/postgresql/16/main/postgresql.conf /etc/postgresql/16/main/postgresql.conf.backup

sudo cp config.conf /etc/postgresql/16/main/postgresql.conf

sudo systemctl restart postgresql 
