1.org.example.config.Config类下的url配置需使用域名
并且将IP   域名的映射加入hosts文件中
2.目前org.example.chaincode.invocation包下的InvokeChaincode
以及QueryCar已经调通
3.数据上链需要与背书策略匹配 否则上链失败
4.项目打包  cd java =》mvn clean package -Dmaven.test.skip=true 会在targe目录下生成XXX.jar
5.项目启动  java -jar XXX.jar
6.后续会使用docker启动