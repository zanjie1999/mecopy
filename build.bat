rd /s /q build
mkdir build
SET CGO_ENABLED=1
SET GOARCH=amd64
SET GOOS=windows
go build -ldflags="-w -s" -o build\mecopy.exe
SET GOARCH=386
go build -ldflags="-w -s" -o build\mecopy-i386.exe
SET CGO_ENABLED=0
SET GOOS=linux
go build -ldflags="-w -s" -o build\mecopy-linux-i386
SET GOARCH=amd64
go build -ldflags="-w -s" -o build\mecopy-linux
SET GOARCH=arm
go build -ldflags="-w -s" -o build\mecopy-linux-arm
SET GOARCH=mips
go build -ldflags="-w -s" -o build\mecopy-linux-mips
SET GOARCH=arm64
go build -ldflags="-w -s" -o build\mecopy-linux-arm64
SET CGO_ENABLED=1
SET GOOS=darwin
go build -ldflags="-w -s" -o build\mecopy-darwin-arm64
SET GOARCH=amd64
go build -ldflags="-w -s" -o build\mecopy-darwin
SET GOOS=freebsd
go build -ldflags="-w -s" -o build/mecopy-freebsd