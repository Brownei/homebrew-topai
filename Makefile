# Makefile
.PHONY: release

release:
	@echo "Releasing version $(VERSION)"
	# Update version in your code if needed
	# Then run GoReleaser
	goreleaser release --clean

build:
	go build -o $(BINARY_NAME) .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-x64 .
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 .

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-macos-x64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-macos-arm64 .

install: build
	sudo install -m 755 $(BINARY_NAME) /usr/local/bin/

clean:
	rm -f $(BINARY_NAME)-*
