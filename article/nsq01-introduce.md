NSQ 实时分布式消息平台
### 简单介绍
官网地址 [https://nsq.io](https://nsq.io)

主要特性
- 分布式：去中心化，无单点故障；
- 可伸缩：支持水平扩展，没有中心化的 broker，支持pub-sub，load-balanced 的消息传递；
- 运维友好：非常容易配置和部署，内置 admin UI；
- 集成方便：目前主流语言都有相应的[客户端支持](https://nsq.io/clients/client_libraries.html) 。

### 快速部署
- nsqd
  - 负责接收、排队、转发消息到客户端的守护进程，可以单独运行
  - 监听 4150（TCP）、4151（HTTP）、4152（HTTPS，可选）端口
- nsqlookupd 
  - 管理拓扑信息
  - 客户端查询 `nsqlookupd` 发现指定 topic 的 `nsqd` 生产者
  - `nsqd` 向 `nsqlookupd` 广播 topic 和 channel 信息
  - 监听 4160（TCP） 和 4161（HTTP）端口；
- nsqadmin

### 功能演示


### 参考资料
- NSQ官网 [https://nsq.io](https://nsq.io)
- 小米信息部技术团队 [走进 NSQ 源码细节](https://xiaomi-info.github.io/2019/12/06/nsq-src/)
