# docker-inspect-2-command-line
A Go lang script that generates the command line to start a docker image using the docker inspect of an existing image

## Usage
Save the docker inspect of you image in a text file (ex.: inspect.txt) and than run
```
docker2go.exe inspect.txt
```

Alternatively, you can run the source code:
```go
go run docker2go.go inspect.txt
```