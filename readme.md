# 我的边车  
利用边车模式我们可以简化微服务的开发，诸多开源项目都在利用这一模式做事情，例如 Istio，微软的 Dapr，都是如此。这里我将收集一些自己认为有意义的场景，做到一个自己的边车中去，积少成多，慢慢来吧。  

## 一个可签发x509证书的CA  
这个不大像是一个边车该做的事情，先放在这里，后续再考虑移走。目前这个CA比较简陋，没有CRL等机制，只能签发。用到了gRPC技术，也通过mTLS加固了gRPC服务器。  

### 下载编译  
0. 自行安装Golang最新版，并配置好Path和GOPATH等环境变量  
1. git clone 代码库  
2. 编译  
```bash
cd <the source code foler>
go build .
```
之后可以在文件夹下看到一个“sidecar"可执行程序  
3. （可选）gRPC 代码生成  
如果想试试gRPC代码生成，可以在装了相应插件后，运行如下命令：  
```bash  
cd pkg/grpc  
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative service.proto  
```
这会在同文件夹下生成service.pb.go和service_grpc.pb.go文件  

### 开发用到的工具和库  
1. cobra命令行工具  
```bash
go install github.com/spf13/cobra-cli@latest  
```
2. gRPC compiler  
[下载地址](https://github.com/protocolbuffers/protobuf/releases)，解压然后加入Path环境变量中  
3. gRPC Go语言代码生成插件  
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest  
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest  

### 命令参数  
可以通过如下命令获取支持的参数：  
```bash
./sidecar -h
./sidecar caserver -h
./sidecar grpcclient -h
```
1. ./sidecar caserver --grpc=<true or false> --mtls=<true or false>
这时启动这个CA server，默认是用http服务器，加了grpc=true的话是用gRPC服务器；当启用gRPC时，有mtls参数可用，用于决定是否用mTLS加固  

2. ./sidecar grpcclient --certid=<id of the signed certificate>
纯粹是为了通过Go程序来访问gRPC Server而做的一个小客户端，通过参数certid给出刚刚生成的cert id，就是调用SignCert服务所返回的值，然后可以得到证书  
