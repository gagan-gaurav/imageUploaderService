# Building Instructions for Windows

Get the tool for zipping the build

```bash
go.exe install github.com/aws/aws-lambda-go/cmd/build-lambda-zip@latest
```

Setting enviornment variables + build command

```bash
$env:GOOS = "linux"; $env:GOARCH = "amd64"; $env:CGO_ENABLED= "0"; go build -o build/main cmd/main.go  
```

Go zipping command (make sure GOPATH is set in environment variables.)

```bash
build-lambda-zip.exe -o build/main.zip build/main
```

