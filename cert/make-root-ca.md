# 生成根证书，一个自签名证书  
一个证书签发机构需要一个可以用于为别人签名的证书，也叫“根证书”。如果不是证书签发代理机构，那么这个证书一定是一张自签名证书，我们可以用如下命令制作一张自签名证书，作为本机构的根证书。注意替换命令中的占位符为实际信息。  
## 准备工作  
```bash
mkdir cert  
mkdir cert\rootCA  
mkdir cert\csr  
```  
## 第一步：生成私钥  
```bash
openssl genrsa -des3 -out rootCA\ca.private.key 2048  
```  
需要你设一个保护该私钥的密码  
## 第二步：生成签名申请  
```bash
openssl req -new -key .\rootCA\ca.private.key -out .\csr\root.csr -subj "/C=CN/O=<你的机构名称>/OU=<机构内部门名称>/CN=<证书所有者的名字，例如域名>/emailAddress=<联系人邮箱>"  
```  
由于使用了private key，所以需要你输入上一步所设置的密码  
## 第三步：对申请自签名，生成自签名证书  
```bash
openssl x509 -req -days 365 -signkey .\rootCA\ca.private.key -in csr\root.csr -out rootCA\root.crt  
```  
## 一些有用的命令  
```bash  
openssl rsa -in .\rootCA\ca.private.key -text  
openssl pkey -in .\rootCA\ca.private.key -text  
```  
第一条命令查看生成的rsa 私钥内容  
第二条命令从私钥中提取rsa 公钥并打印  

···bash
openssl x509 -in rootCA\root.crt -noout -text
···  
查看x509证书内容  

