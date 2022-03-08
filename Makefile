# install developer tools
.PHONY: install-tools
install-tools:
	go install github.com/cosmtrek/air
	go install golang.org/x/tools/cmd/goimports


.PHONY: fmt
fmt:
	goimports -l -w $$(find . -name "*.go")
