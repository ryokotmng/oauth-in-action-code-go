# install developer tools
.PHONY: install-tools fmt
install-tools:
	go install github.com/cosmtrek/air
	go install golang.org/x/tools/cmd/goimports
	docker pull redis

fmt:
	goimports -l -w -local "github.com/ryokotmng/oauth-in-action-code-go" $$(find . -name "*.go")

start-redis:
	docker run --name redis-oauth -p 6379:6379 -d redis

stop-redis:
	docker stop redis-oauth