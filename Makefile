APP_NAME=mdx
MAIN=main.go
BIN_DIR=bin
DIST_DIR=bin
NFPM?=nfpm
GOARCH?=amd64
MDX_VERSION?=v9.9.9

.PHONY: run build build-linux build-macos-amd64 build-macos-arm64 build-windows-amd64 build-windows-arm64 install clean build-linux-bin package-deb package-rpm package-apk package-archlinux package-linux

run:
	go run $(MAIN)

build:
	go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN)

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-linux $(MAIN)

build-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-macos-amd64 $(MAIN)

build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-macos-arm64 $(MAIN)

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN)

build-windows-arm64:
	GOOS=windows GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-arm64.exe $(MAIN)

build-linux-bin:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN)

package-deb: build-linux-bin
	MDX_VERSION=$(MDX_VERSION) GOARCH=$(GOARCH) $(NFPM) package --packager deb --config $(CURDIR)/packaging/deb.yaml --target $(DIST_DIR)/

package-rpm: build-linux-bin
	MDX_VERSION=$(MDX_VERSION) GOARCH=$(GOARCH) $(NFPM) package --packager rpm --config $(CURDIR)/packaging/rpm.yaml --target $(DIST_DIR)/

package-apk: build-linux-bin
	MDX_VERSION=$(MDX_VERSION) GOARCH=$(GOARCH) $(NFPM) package --packager apk --config $(CURDIR)/packaging/apk.yaml --target $(DIST_DIR)/

package-archlinux: build-linux-bin
	MDX_VERSION=$(MDX_VERSION) GOARCH=$(GOARCH) $(NFPM) package --packager archlinux --config $(CURDIR)/packaging/archlinux.yaml --target $(DIST_DIR)/

package-linux: package-deb package-rpm package-apk package-archlinux

install:
	go install

clean:
	rm -rf bin
