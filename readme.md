# Start server
```go
go run server/grpc/main.go
go run server/rest/*.go
```
# Kafka
## Start zookeeper
```zsh
bin/zookeeper-server-start.sh config/zookeeper.properties
```
## Start server/broker
```zsh
bin/kafka-server-start.sh config/server.1.properties
bin/kafka-server-start.sh config/server.2.properties
bin/kafka-server-start.sh config/server.3.properties
```
## List topic
```zsh
bin/kafka-topics.sh --bootstrap-server localhost:9094 --list
```
## Delete topic
```zsh
bin/kafka-topics.sh --delete --bootstrap-server localhost:9093 --topic <topic name>
```
## Create topic
```zsh
bin/kafka-topics.sh --create --topic mytopic --bootstrap-server localhost:9094 --partitions 4 --replication-factor 2
```
## Describe topic
```zsh
bin/kafka-topics.sh --describe --topic mytopic --bootstrap-server localhost:9095
```
## Test
```zsh
go run kafka/main.go
```