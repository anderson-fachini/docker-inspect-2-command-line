rm docker2go.windows-amd64.exe
rm docker2go.linux-amd64

set GOARCH=amd64
set GOOS=windows
go build docker2go.go
ren docker2go.exe docker2go.windows-amd64.exe

set GOOS=linux
go build docker2go.go
ren docker2go docker2go.linux-amd64

pause