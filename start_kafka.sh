# !/bin/bash

# Start Zookeeper
echo 'Starting Zookeeper'
./kafka/bin/zookeeper-server-start.sh kafka/config/zookeeper.properties &

# Start Broker 0
echo 'Starting Broker 0'
./kafka/bin/kafka-server-start.sh kafka/config/server-0.properties &

# Start Broker 1
echo 'Starting Broker 1'
./kafka/bin/kafka-server-start.sh kafka/config/server-1.properties &

# Start Broker 2
echo 'Starting Broker 2'
./kafka/bin/kafka-server-start.sh kafka/config/server-2.properties
