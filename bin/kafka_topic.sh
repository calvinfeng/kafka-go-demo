# !/bin/bash

# Try to delete the topic first
./kafka/bin/kafka-topics.sh --delete --zookeeper localhost:2181 --topic chat
./kafka/bin/kafka-topics.sh --list --zookeeper localhost:2181

# Create a topic named chat
./kafka/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 3 --partitions 1 --topic chat
./kafka/bin/kafka-topics.sh --describe --zookeeper localhost:2181 --topic chat
