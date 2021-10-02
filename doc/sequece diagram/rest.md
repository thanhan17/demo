@startuml
title rest
participant User        as user
control     Handler     as gin
participant UserService as us
participant gRPC        as grpc
database    Cache       as cache
database    MySQL       as db

user -> gin: send request
activate gin

gin -> us: handle request
activate us

us -> grpc: send request
activate grpc

group redsync(Create)
grpc -> cache: increase & read id counter
activate cache

cache -> grpc: return value
deactivate cache
end

group mysql
grpc -> db: CRUD database
activate db

db -> grpc: result, error
deactivate db
end

grpc -> us: return response 
deactivate grpc

us -> gin: return result
deactivate us

gin -> user: return response
deactivate gin
@enduml