### fork 源码
访问 [GitHub NSQ源码](https://github.com/nsqio/nsq) ，为了以后方便代码管理，我们可以 fork nsq 源码到自己的 GitHub 帐号仓库下，比如我的 https://github.com/geekymv/nsq

### 下载源码
下载代码到本地机器，访问自己的 GitHub 帐号刚刚 fork 点击 nsq 源码，点击 Code -> SSH -> 复制地址
在自己本地机器执行 `git clone` 下载源码
```shell
git clone git@github.com:geekymv/nsq.git
```

创建 git 分支
```shell
git checkout -b nsq-annotated
```
这样我们以后看代码都在这个分支上进行，可以任意修改代码、添加注释等
```shell
git push origin nsq-annotated
```

### 导入到开发工具
nsq 源码的依赖是通过 `go mod` 管理的，关于 `go mod` 的使用不熟悉的朋友可以自己搜索资料学习下，我这里就不过多赘述了。
将下载下来的源码导入到自己熟悉的开发工具，我这里导入到 GoLand。

### 下载依赖
```shell
go mod tidy
```




### NSQ Design

一个 `nsqd` 实例处理大量的数据流，这些数据流被称为 `topics`，一个 `topic` 有一个或多个 `channels`，`topic` 会将消息拷贝到与之关联的 `channel` 中，
通常，一个 `channel` 对应一个下游消费 `topic` 的服务。这个服务通常会部署多个实例，`channel` 中的数据会随机的发给些实例中的一个，也就是说每个实例处理这个 `channel` 中的部分数据。

topic 和 channel 不是预先创建的，
生产者与 nsqd 建立连接，然后向一个指定名称的 topic 发布消息的时候，nsqd 判断如果这个 topic 还不存在则会创建一个 topic，
消费者与 nsqd 建立连接，消费指定名称的topic/channel，channel 会被创建， topic 如果不存在的话也会被创建。

nsqd 启动的时候与 nsqloopupd 建立连接，并将自己的 topic/channel 信息同步给 nsqlookupd。
消费者 从 nsqlookupd 中查找提供自己订阅的 topic 的 nsqd 的地址，这种方式将生产者和消费者解耦，降低系统复杂性。
生产者只与 nsqd 建立连接，发送消息到指定的 topic，它不与 nsqlookupd 通信。

nsqd 与 nsqd 之间是独立的，不需要通信或协调， nsqlookupd 之间也是如此。