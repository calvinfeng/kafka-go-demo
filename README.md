# Kafkapo (Golang) - Realtime Message Application
## Setup for Ubuntu
### Installing default JRE/JDK
First of all, we need Java to run Kafka.
```shell
sudo apt-get update
```

Check if Java is already installed, I am assuming your machine is brand new
```shell
java -version # => The program java can be found in the following packages...
```

Run `apt-get`
```shell
sudo apt-get install default-jre
```
```shell
sudo apt-get install default-jdk
```
That's all we need for Java!

### Installing Golang Client
Assuming that you have Golang installed on your system, the next thing we need is `librdkafka-dev` which is a C/C++ library for Kafka. We will install `librdkafka` through `apt` using Confluent's Debian repository.

First install Confluent's public key
```shell
wget -qO - http://packages.confluent.io/deb/3.3/archive.key | sudo apt-key add -
```

Add the repository to your `etc/apt/sources.list`
```shell
sudo add-apt-repository "deb [arch=amd64] http://packages.confluent.io/deb/3.3 stable main"
```

Run `apt-get install`
```shell
sudo apt-get install librdkafka-dev
```

Then finally, we can do a `go get`
```shell
go get -u github.com/confluentinc/confluent-kafka-go/kafka
```

The Golang client will show up in your `$GOPATH/src/github.com/`

## How-to Kafka
The shell script `start_kafka.sh` has the required commands to start the Kafka server. We wiill go through them line
by line here.

### Servers
Starts Zookeeper
```shell
kafka/bin/zookeeper-server-start.sh config/zookeeper.properties
```

Starts our first server-0
```shell
kafka/bin/kafka-server-start.sh config/server-0.properties
```

### Topics
Create an example topic
```shell
kafka/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic example
```

We can list all the topics with this command
```shell
 kafka/bin/kafka-topics.sh --list --zookeeper localhost:2181
 ```

### Produce Messages
Kafka comes with a command line client that will take inputs from command line inputs and send them out as messages to the Kafka clusters.
```shell
kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic example
```

## Start the Golang server
First of all, add this project into your `GOPATH` and include GOPATH in your path
```
export GOPATH=/path/to/Kafkapo
export PATH=$PATH:$GOPATH/bin
```

And then install Gorilla websocket & Kafka Golang client
```
cd /path/to/Kafkapo
go get github.com/gorilla/websocket
go get github.com/confluentinc/confluent-kafka-go/kafka
```

And then `cd Kafkapo/src/server` and run `go install && server` to start the server
