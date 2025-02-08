test::
	go build -trimpath -ldflags '-s -w' -v -o D:\Services\bin\filecompact.exe ./

wsl::
	wsl --exec go build -trimpath -ldflags '-s -w' -v -o bin/filecompact ./

all:: test wsl

build::
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o bin/filecompact-linux-amd64 ./
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '-s -w' -o bin/filecompact-linux-arm64 ./
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags '-s -w' -o bin/filecompact-linux-armv7 ./
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o bin/filecompact-amd64.exe ./
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags '-s -w' -o bin/filecompact-arm64.exe ./
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GOAMD64=v3 go build -trimpath -ldflags '-s -w' -o bin/filecompact-arm64-v3.exe ./
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o bin/filecompact-darwin-amd64 ./
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags '-s -w' -o bin/filecompact-darwin-arm64 ./
 
