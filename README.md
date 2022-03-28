# OAuth in Action in Go
This repository is my try to translate [OAuth 2 in Action](https://github.com/oauthinaction/oauth-in-action-code) source code into Go.

![Cover of OAuth 2 in Action](https://images.manning.com/255/340/resize/book/e/14336f9-6493-46dc-938c-11a34c9d20ac/Richer-OAuth2-HI.png)

https://www.manning.com/books/oauth-2-in-action

## Requirement
- Go 1.17

## Libraries
- WAF: [Gin](https://github.com/gin-gonic/gin)
  - [CORS gin's middleware](https://github.com/gin-contrib/cors)
- DB: [Redis](https://github.com/go-redis/redis)
- Others:
  - [oauth2](https://pkg.go.dev/golang.org/x/oauth2)
    - NOTE: To understand detailed internal implementation, this repository does not use some useful features of this package

## Tools
- File Watcher: [Air](https://github.com/cosmtrek/air)

Execute 'make install-tools' at the root directory to install necessary tools.  

## Setup
Install useful tools
```sh
$ make install-tools
```

Start up redis
```sh
$ make start-redis
```

## How to Use
The folder structure is quite similar to [OAuth 2 in Action](https://github.com/oauthinaction/oauth-in-action-code), but the entry points for authorization, client and protected resource servers are under each folder.  
You should execute 'air' of 'go run main.go' command under each folder in [cmd](https://github.com/ryokotmng/oauth-in-action-code-go/tree/main/cmd).  

Ports are all same with the original repository, except for redis which replace over nosql.

| server | port |  
| -- | -- |  
| authorization | 9001 |  
| client | 9000 |  
| protected resource | 9002 |  
| redis | 6379 |  
