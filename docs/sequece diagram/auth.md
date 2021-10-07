@startuml
title Authenticate
participant User        as user
control     Gin         as gin
participant Auth        as auth
database    Cache       as cache
user -> gin: request
activate gin
gin -> auth: authenticate
activate auth
auth ->  cache: save or read token
activate cache
cache -> auth: return token
deactivate cache
auth -> gin: return auth result
deactivate auth
gin -> user: return json result
deactivate gin
@enduml