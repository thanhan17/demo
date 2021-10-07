@startuml
title Login
participant User        as user
participant Middleware  as mdw
control     Handlers    as handler
participant Token       as token
participant JWT         as jwt
participant Cache       as cache
user -> mdw: request logout
activate mdw
mdw -> token: TokenValid(*http.Request)
activate token
token -> mdw: return error
deactivate token
mdw -> handler: next
deactivate mdw
activate handler
group logout
handler -> token: ExtractTokenMetadata(*http.Request)
activate token
group ExtractTokenMetadata
token -> jwt: jwt.Parse(tokenString)
activate jwt
jwt -> token: return (*jwt.Token, error)
deactivate jwt
token -> jwt: extract(token *jwt.Token)
activate jwt
jwt -> token: return (*AccessDetails, error)
deactivate jwt
end
token -> handler: return (*AccessDetails, error)
deactivate token
handler -> cache: DeleteTokens(context.Context, *AccessDetails)
activate cache
cache -> handler: return error
deactivate cache
end
handler -> user: response
deactivate handler
@enduml