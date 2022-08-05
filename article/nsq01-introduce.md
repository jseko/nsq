NSQ 实时分布式消息平台

### 为什么要选择学习 NSQ 源码
已经学习了 Go 语言基础，但是不知道怎么去使用，那么可以去看看一些开源框架里面都是怎么去使用这些技术点的，
不仅仅学习技术点，还可以学习一些框架的设计，这样对我们开发水平的提高是有非常大的好处的。

- NSQ 使用 Go 语言开发
- NSQ 设计到的技术点
  - slice、map、struct、interface、error
  - goroutine  
  - channel
  - lock
  - 网络 tcp、http
  - 自定义数据报文、编解码
  - 其他

### 简单介绍
官网地址 [https://nsq.io](https://nsq.io)

主要特性
- 分布式：去中心化，无单点故障；
- 可伸缩：支持水平扩展，没有中心化的 broker，支持pub-sub，load-balanced 的消息传递；
- 运维友好：非常容易配置和部署，内置 admin UI；
- 集成方便：目前主流语言都有相应的[客户端支持](https://nsq.io/clients/client_libraries.html) 。

### 核心组件
- nsqd
  - 负责接收、排队、转发消息到客户端的守护进程，可以单独运行
  - 监听 4150（TCP）、4151（HTTP）、4152（HTTPS，可选）端口
- nsqlookupd
  - 管理拓扑信息
  - 客户端查询 `nsqlookupd` 发现指定 topic 的 `nsqd` 生产者
  - `nsqd` 向 `nsqlookupd` 广播 topic 和 channel 信息
  - 监听 4160（TCP） 和 4161（HTTP）端口；
- nsqadmin
  - 提供一个 web ui, 用于实时查看集群信息，进行各种任务管理。



### 快速部署
准备 3 台机器
```shell
docker run -id --name node01 -p 4160:4160 -p 4161:4161 centos:7
```

| 主机名  | IP地址         | 组件                              |
| ------- | :------------- | :-------------------------------- |
| node101 | 192.168.56.101 | nsqd, nsqlookupd  |
| node102 | 192.168.56.102 | nsqd, nsqlookupd |
| node103 | 192.168.56.103 | nsqd, nsqlookupd, nsqadmin |

从 nsq [官网下载](https://nsq.io/deployment/installing.html) 这里演示在 Linux 系统安装过程
- 下载 nsq-1.2.1.linux-amd64.go1.16.6.tar.gz
- 解压
```shell
tar -zxvf nsq-1.2.1.linux-amd64.go1.16.6.tar.gz
mv nsq-1.2.1.linux-amd64.go1.16.6 nsq-1.2.1

cd nsq-1.2.1/bin
```
- 启动 nsqlookupd
  ```shell
  $ nsqlookupd
  ```
- 启动 nsqd
  ```shell
  $ nsqd --lookupd-tcp-address=127.0.0.1:4160
  ```
- 启动 nsqadmin
  ```shell
  $ nsqadmin --lookupd-http-address=127.0.0.1:4161
  ```

### 功能演示


### 参考资料
- NSQ官网 [https://nsq.io](https://nsq.io)