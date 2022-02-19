# OAuth in Action in Go
This repository is my try to translate [OAuth 2 in Action](https://github.com/oauthinaction/oauth-in-action-code) source code into Go.

## Requirement
- Go 1.17

## Libraries Used
- WAF: [Gin](https://github.com/gin-gonic/gin)
- File Change Watcher: [Air](https://github.com/cosmtrek/air)
- DB: [Redis](https://github.com/go-redis/redis)

## How to Use
The folder structure is quite ismilar to [OAuth 2 in Action](https://github.com/oauthinaction/oauth-in-action-code), but the entry points for authorization server, client and protected resource are under each folder.  
You should execute 'air' of 'go run main.go' command under each folder to start the server.  

Ports are all same with the original repository.


| server | port |  
| -- | -- |  
| authorization | 9001 |  
| client | 9000 |  
| protected resource | 9002 |  

