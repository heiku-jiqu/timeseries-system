#!/bin/bash

kafka-topics --create \
  --topic coinbase-ticker \
  --bootstrap-server broker:29092 \
  --partitions 2 --replication-factor 1
