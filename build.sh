cd `dirname $0`
mkdir -p build
rm -rf build/*
export CGO_ENABLED=0
export GOARCH=amd64
export GOOS=windows
go build -ldflags="-w -s" -o build/mecopy.exe
export GOARCH=386
go build -ldflags="-w -s" -o build/mecopy-i386.exe
export CGO_ENABLED=0
export GOOS=linux
go build -ldflags="-w -s" -o build/mecopy-linux-i386
export GOARCH=amd64
go build -ldflags="-w -s" -o build/mecopy-linux
export GOARCH=arm
go build -ldflags="-w -s" -o build/mecopy-linux-arm
export GOARCH=mips
go build -ldflags="-w -s" -o build/mecopy-linux-mips
export GOARCH=arm64
go build -ldflags="-w -s" -o build/mecopy-linux-arm64
export GOOS=darwin
go build -ldflags="-w -s" -o build/mecopy-darwin-arm64
export GOARCH=amd64
go build -ldflags="-w -s" -o build/mecopy-darwin
export GOOS=freebsd
go build -ldflags="-w -s" -o build/mecopy-freebsd