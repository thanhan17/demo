@startuml
title Logging
participant User        as server
queue       Kafka       as kafka
collections Logs        as log
server -> kafka: send stream log
kafka -> log: receive log and write to file
@enduml