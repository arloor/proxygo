SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build -o build/proxygo_mac

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o build/proxygo

go build -o build/proxygo_dev.exe

go build -ldflags="-H windowsgui" -o build/proxygo.exe