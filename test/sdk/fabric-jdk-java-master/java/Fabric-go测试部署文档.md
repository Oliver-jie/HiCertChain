## 1、系统环境的准备

```
# 更新apt包： 
apt-get update

# 安装vim:apt install vim

#apt加速（https://mirrors.tuna.tsinghua.edu.cn/help/ubuntu/）
vim /etc/apt/sources.listapt-get update

# 安装wget: 
apt install wget

# 安装Cmake:
wget https://cmake.org/files/v3.18/cmake-3.18.0-rc1-Darwin-x86_64.tar.gz
tar xzvf cmake-3.17.1-Linux-x86_64.tar.gz
mv cmake-3.17.1-Linux-x86_64 /opt/cmake-3.17.1
ln -sf /opt/cmake-3.17.1/bin/* /usr/bin/cmake --version

#6) 安装make: 
apt-get install make

```

## 2、配置Fabric开发环境

```
##参考文献：https://hyperledger-fabric.readthedocs.io/en/latest/dev-setup/devenv.html

#安装Git：
apt-get install git

#安装 Go version 1.14.x：
wget https://gomirrors.org/dl/go/go1.14.4.linux-amd64.tar.gztar -xzf go1.8.linux-amd64.tar.gz  -C /usr/local/usr/local/go/bin/go version

#切换7牛云go module镜像
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

# 设置环境变量：
vim ~/.bashrc

    export GOPATH=/opt/go
    export GOROOT=/usr/local/go
    export GOARCH=386
    export GOOS=linux
    export GOBIN=$GOROOT/bin/
    export GOTOOLS=$GOROOT/pkg/tool/
    export PATH=$PATH:$GOBIN:$GOTOOLS

#安装jq
sudo apt-get install jq

#安装SoftHSM：
sudo apt install softhsm

```

## 3、Fabric的安装

参考文献：https\://hyperledgercn.github.io/hyperledgerDocs/getting_started/

### 3.1 安装Samples, Binaries & Docker Images

```
参考：https://hyperledger-fabric.readthedocs.io/en/release-2.1/install.html
下载https://hyperledger-fabric.readthedocs.io/en/release-2.1/install.html把这个文件存为文件：https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh

```

#### **3.1.1 自动安装**

```
$ ./bootstrap.sh
bootstrap.sh会自动下载binaries & docker images.

```

#### **3.1.2 手动安装**

> **1.下载如下的tar包：**

```
https://github.com/hyperledger/fabric/releases/download/v2.1.1/hyperledger-fabric-linux-amd64-2.1.1.tar.gz
https://github.com/hyperledger/fabric-ca/releases/download/v1.4.7/hyperledger-fabric-ca-linux-amd64-1.4.7.tar.gz

```

> **2. 解压缩**

```
tar xzvf hyperledger-fabric-linux-amd64-2.1.1.tar.gz
tar xzvf hyperledger-fabric-ca-linux-amd64-1.4.7.tar.gz

```

> **3 .(把二进制bin文件copy到/fabric-samples/bin目录下.**

```
mv ./bin /usr/local/src/fabric-samples 
mv ./config /usr/local/src/fabric-samples 

```

> **4. 配置/etc/profile**

```
FABRIC_ROOT=/usr/local/src/fabric-samples
export GOROOT=/usr/local/go              # 安装目录。
export GOPATH=$HOME/go     # 工作环境
export GOBIN=$GOROOT/bin           # 可执行文件存放
export PATH=/usr/local/src/fabric/bin:$PATH
export PATH=/usr/local/go/bin:$PATH
export FABRIC_CFG_PATH=$FABRIC_ROOT/config/
export GOPROXY=https://goproxy.cn

# Environment variables for Org1
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$FABRIC_ROOT/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=$FABRIC_ROOT/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
#export FABRIC_LOGGING_SPEC=DEBUG

# Environment variables for Org2
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
#Anaconda
export PATH=$PATH:/home/zjj/anaconda3/bin
export CORE_PEER_ADDRESS=localhost:9051

```

### 3.2 测试网络

参考文献：<https://hyperledger-fabric.readthedocs.io/en/release-2.1/test_network.html>

> 1.进入工作目录 

```
cd /usr/local/src/fabric-samples/test-network

```

> 2\. 切换到root账户：

```
sudo su

```

> 3\. 开启网络

```
$./network.sh up

```

> 4.创建通道

```
./network.sh createChannel -c xxribd

```

> 5.部署智能合约

```
./network.sh deployCC -c xxribd

```

> 6.查询链上数据

```
peer chaincode query -C xxribd -n fabcar -c '{"Args":["queryCar","CAR9"]}'

```

> 7.更新链上数据

```
peer chaincode invoke 
-o localhost:7050 
--ordererTLSHostnameOverride orderer.example.com 
--tls 
--cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem 
-C xxribd
-n fabcar 
--peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt 
--peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt 
-c '{"function":"changeCarOwner","Args":["CAR9","Dave"]}'

```

> 8.关闭网络

```
sudo ./network.sh down

```


