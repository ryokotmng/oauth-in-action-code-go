# install developer tools
.PHONY: install-tools
install-tools:
	go install github.com/cosmtrek/air
	go install golang.org/x/tools/cmd/goimports


.PHONY: fmt
fmt:
	goimports -l -w -local "github.com/ryokotmng/oauth-in-action-code-go" $$(find . -name "*.go")
