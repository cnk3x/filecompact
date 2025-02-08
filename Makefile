build::
	go build -trimpath -ldflags '-s -w' -v -o bin/filecompact.exe ./

install::
	go build -trimpath -ldflags '-s -w' -v -o D:\Services\bin\filecompact.exe ./

wsl::
	wsl --exec go build -trimpath -ldflags '-s -w' -v -o bin/filecompact ./

all:: install wsl build
