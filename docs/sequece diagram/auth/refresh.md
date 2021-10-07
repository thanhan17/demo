@startuml
title Login
participant User        as user
control     Handlers    as handler
participant Token       as token
participant JWT         as jwt
participant Cache       as cache
user -> handler: request refresh
activate handler
group Refresh
handler -> jwt: Parse tokenString
activate jwt
jwt -> handler: return {refresh_uuid, user_id}
deactivate jwt
handler -> cache: DeleteRefresh(ctx context.Context, refreshUuid string)
activate cache
cache -> handler: return error
deactivate cache
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
handler -> cache: CreateAuth(ctx, userId, *TokenDetails)
activate cache
cache -> handler: return error
deactivate cache
end
handler -> user: return {access_token, refresh_token}
deactivate handler
@enduml