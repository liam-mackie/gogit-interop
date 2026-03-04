LIB_NAME = libgogit
SRC = .

.PHONY: generate darwin-arm64 darwin-amd64 linux-amd64 windows-amd64 all clean

generate:
	cd ../generate && go run .

darwin-arm64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -buildmode=c-shared -o $(LIB_NAME).dylib $(SRC)

darwin-amd64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build -buildmode=c-shared -o $(LIB_NAME)_amd64.dylib $(SRC)

linux-amd64:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc \
		go build -buildmode=c-shared -o $(LIB_NAME).so $(SRC)

windows-amd64:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
		go build -buildmode=c-shared -o $(LIB_NAME).dll $(SRC)

all: darwin-arm64 darwin-amd64

clean:
	rm -f $(LIB_NAME).dylib $(LIB_NAME)_amd64.dylib $(LIB_NAME).so $(LIB_NAME).dll $(LIB_NAME).h
