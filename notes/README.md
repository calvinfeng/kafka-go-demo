# Kafka
## Terminology
* topics
* brokers
* producers
* consumers

Each node in the cluster is a broker. Each broker can have multiple partitions.

## Topics
Topics are divided into partitions which allow you to parallelize a topic by splitting data across multiple brokers. Each
partition can be placed on a separate machine to allow for multiple consumers to read from a topic in parallel.

Each message within a partition has an ID called its **offset**. Consumers can read messages starting from a specific
offset and are allowed to read from any offset point they choose.

Each message in a Kafka cluster can be uniquely identified by a tuple, (topic, partition, offset)

Example: **Laser topic**

|              | Offset 0 | Offset 1 | ... |
| ------------ | -------- | -------- | --- |
| Partition 0  | msg      | msg      | ... |
| Partition 1  | msg      | msg      | ... |

Kafka does not keep state on what the consumers are reading from a topic

## Brokers & Partitions
Each broker holds a number of partitions and each of these partitions can be either a leader or a replica for a topic.

| Broker 1       | Broker 2    | Broker 3    |
| ----------     | ----------- | ----------- |
| **Partitio 0** | Partition 0 | Partition 0 |
| Partitio 1 | **Partition 1** | Partition 1 |
| Partitio 2 | Partition 2 | **Partition 2** |

**Bold** partition is the leader, others are replicas.

## Consistency and Availability
1. Messages sent to a topic partition will be appended to the commit log in the order they are sent.
2. A single consumer instance will see messages in the order they appear in the log.
3. A message is committed when all in-sync replicas have applied it to their log.
4. Any committed message will not be lost, as long as at least one in-sync replica is alive.

### Handling Failures During Writes
When a replica fails, writes will no longer reach the failed replica and it will no longer receive messages. As time goes on, it will fall further and further out of sync with the leader.

When the second replica fails, it will no longer receive messages and it too becomes out of sync with the leader.

When the third replica fails, that is the leader also fails, we are basically left with three dead replicas.

The third replica is actually still in sync. It cannot receive any new data but it is in sync with everything that was possible to receive, while the second replica is missing some data and the first replica is missing even more data. We have (2) possible solutions in this scenario.

1. Wait until the leader is back up before continuing. Once the leader is back up, it will begin receiving and writing messages and as the replicas are brought back online they will be made in sync with the leader.

2. Elect the first broker to come back up as the new leader. This broker will be out of sync with the original leader. As additional brokers come back up, they will see that they have committed messages that do not exist on the new leader and drop those messages.

With option 2, we minimize the downtime but at the cost of losing the committed messages.

With option 1, we minimize the loss of the committed messages but it takes us longer to get the system back up.
