@startuml
title Login
participant User        as user
control     Handlers    as handler
participant Token       as token
participant JWT         as jwt
participant Cache       as cache
user -> handler: request login
activate handler
group login
handler -> token: CreateToken(userId)
activate token
group CreateToken
token -> jwt: Create jwt {access_uuid, user_id, exp}
activate jwt
jwt -> token: return jwt
deactivate jwt
token -> jwt: Create jwt {refresh_uuid, user_id, exp}
activate jwt
jwt -> token: return jwt
deactivate jwt
end
token -> handler: return (*TokenDetails, error)
deactivate token
handler -> cache: CreateAuth(ctx, userId, *TokenDetails)
activate cache
cache -> handler: return error
deactivate cache
end
handler -> user: return {access_token, refresh_token}
deactivate handler
@enduml