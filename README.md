# Kafkapo (Golang) - Realtime Message Application
## Quick start
If you have everything installed already, just run the following commands

```
./kafka/bin/zookeeper-server-start.sh kafka/config/zookeeper.properties
```

Open another tab
```
./kafka/bin/kafka-server-start.sh kafka/config/server-0.properties
```

Another tab
```
./kafka/bin/kafka-server-start.sh kafka/config/server-1.properties
```

Another tab again
```
./kafka/bin/kafka-server-start.sh kafka/config/server-2.properties
```

## Setup for Ubuntu users
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
That's all we need for Java to start Kafka!

## How-to Kafka
As described in the **Quick start** section, we will walk through the lines in details.

### Servers
Starts Zookeeper
```shell
kafka/bin/zookeeper-server-start.sh kafka/config/zookeeper.properties
```

Starts our first server-0
```shell
kafka/bin/kafka-server-start.sh kafka/config/server-0.properties
```

### Topics
#### Create
Create a topic named `chat` with replication factor of 1 and on partition 1
```shell
kafka/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic chat
```

#### List
We can list all the topics with this command
```
 kafka/bin/kafka-topics.sh --list --zookeeper localhost:2181
 ```

#### Describe
We can look at a topic in detail using `describe`
```
./kafka/bin/kafka-topics.sh --describe --zookeeper localhost:2181 --topic chat
```

#### Delete
We can delete the topic `chat` with this command
```
./kafka/bin/kafka-topics.sh --delete --zookeeper localhost:2181 --topic chat
```

### CLI Producer
Kafka comes with a command line client that will take inputs from command line inputs and send them out as messages to the Kafka clusters.
```
./kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic chat
<Enter your messages>
<Enter again>
<again...>
```

### CLI Consumer
Similarly we can start another CLI on another tab to listen to the messages produced on a particular topic.
```
./kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic chat --from-beginning
```

## Start the Golang server
### Setup Go
First of all, add this project into your `GOPATH` and include GOPATH in your path
```
export GOPATH=/path/to/Kafkapo
export PATH=$PATH:$GOPATH/bin
```

### Installing Golang Client
There are two choices of Golang Client

* `confluentinc/confluent-kafka-go/kafka`
* `Shopify/sarama`

#### Confluent Go
We are going to do the set up for Confluent's Kafka Golang client first. Assuming that you have Golang installed on your system, we need to install `librdkafka-dev` which is a C/C++ library for Confluent Golang client to interact with Kafka. We will install `librdkafka` through `apt` using Confluent's Debian repository.

First install Confluent's public key
```shell
wget -qO - http://packages.confluent.io/deb/3.3/archive.key | sudo apt-key add -
```

Add the repository to your `etc/apt/sources.list`
```shell
sudo add-apt-repository "deb [arch=amd64] http://packages.confluent.io/deb/3.3 stable main"
```

And then run and an update
```shell
sudo apt-get update
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

#### Sarama Go
This one is much easier, simply do a `go get`
```shell
go get github.com/Shopify/sarama
```

### Installing Gorilla dependencies
Also install gorilla, we need it for websocket and routing.
```
cd /path/to/Kafkapo
go get github.com/gorilla/websocket
go get github.com/gorilla/mux
```

### Setup NPM
Then run `npm install` and `npm run build:watch` to compile JavaScript

### Start Server
And then `cd Kafkapo/src` and

* run `go install cgo_server && cgo_server` to start the Confluent Go server
* run `go install sarama_server && sarama_server` to start the Shopify Sarama server

## To-do Features
I need to use `godep`
