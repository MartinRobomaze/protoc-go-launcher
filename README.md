# protoc-go-launcher
Launcher of protoc &amp; grpc Go code generator.
Handles downloading protoc and Go protoc plugins automatically.
### Installation
```shell
go install github.com/MartinRobomaze/protoc-go-launcher@latest
```
### Usage
Command usage:
```shell
protoc-go-launcher --protoc_version <VERSION> PROTOC_COMMANDS
```
Example with protobuf spec in file `helloworld/helloworld.proto`:
```shell
protoc-go-launcher --protoc_version 32.0  \
    --go_out=. --go_opt=paths=source_relative \   
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    helloworld/helloworld.proto
```