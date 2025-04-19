APP_NAME=mdx
MAIN=main.go
BIN_DIR=bin

run:
	go run $(MAIN)

build:
	mkdir $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(MAIN)

build-linux:
	mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-linux $(MAIN)

build-macos-amd64:
	mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-macos-amd64 $(MAIN)

build-macos-arm64:
	mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-macos-arm64 $(MAIN)

build-windows-amd64:
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN)

build-windows-arm64:
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-arm64.exe $(MAIN)

install:
	go install

clean:
	rm -rf bin
