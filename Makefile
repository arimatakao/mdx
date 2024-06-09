APP_NAME=mdx
MAIN=main.go
BIN_DIR=bin

run:
	go run $(MAIN)

build:
	mkdir $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(MAIN)

clean:
	rm -rf bin
