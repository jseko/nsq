通过上一篇文章的分析我们知道 NSQ 使用`go-svc` 库，实现了 `Init`、`Start` 和 `Stop` 三个步骤。

#### Init方法
创建 NSQD 对象，代码如下：
```go
// 创建 nsqd
nsqd, err := nsqd.New(opts)
```

下面是创建 NSQD 对象的核心代码，这里只保留了核心代码
```go
func New(opts *Options) (*NSQD, error) {
	var err error

	// 省略部分代码

	n := &NSQD{
		startTime: time.Now(),
		// topic 存储 map
		topicMap:             make(map[string]*Topic),
		exitChan:             make(chan int),
		notifyChan:           make(chan interface{}),
		optsNotificationChan: make(chan struct{}, 1),
		dl:                   dirlock.New(dataPath),
	}
	// 创建 context
	n.ctx, n.ctxCancel = context.WithCancel(context.Background())
	
	// 省略部分代码

	n.logf(LOG_INFO, version.String("nsqd"))
	n.logf(LOG_INFO, "ID: %d", opts.ID)

	// 创建 tcpServer
	n.tcpServer = &tcpServer{nsqd: n}
	// tcp listener
	n.tcpListener, err = net.Listen("tcp", opts.TCPAddress)
	if err != nil {
		return nil, fmt.Errorf("listen (%s) failed - %s", opts.TCPAddress, err)
	}
	if opts.HTTPAddress != "" {
		// 创建 http listener
		n.httpListener, err = net.Listen("tcp", opts.HTTPAddress)
		if err != nil {
			return nil, fmt.Errorf("listen (%s) failed - %s", opts.HTTPAddress, err)
		}
	}
	if n.tlsConfig != nil && opts.HTTPSAddress != "" {
		n.httpsListener, err = tls.Listen("tcp", opts.HTTPSAddress, n.tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("listen (%s) failed - %s", opts.HTTPSAddress, err)
		}
	}
	
	// 省略部分代码
	
	return n, nil
}
```
NSQD 中的一些核心字段
- topicMap map[string]*Topic： 用来存储 Topic 信息
- tcpServer *tcpServer： tcpServer 实现了 TCPHandler 接口，主要负责出来每个连接的I/O
- tcpListener net.Listener：监听 TCP 连接，生产者和消费者都可以与nsqd建立 TCP 连接
- httpListener / httpsListener net.Listener：提供 HTTP 接口


#### Start 启动 NSQD
```go
func (p *program) Start() error {
	// 省略部分代码
	
	go func() {
		// 启动 nsqd
		err := p.nsqd.Main()
		if err != nil {
			p.Stop()
			os.Exit(1)
		}
	}()

	return nil
}
```

NSQD 中的 Main 方法，根据 tcpListener、httpListener、httpsListener 分别在每个 goroutine 中创建 TCPServer、httpServer 和 httpsServer。
```go
func (n *NSQD) Main() error {
    exitCh := make(chan error)
	
	// 省略部分代码
	
	// tcp
	n.waitGroup.Wrap(func() {
		// 开启一个 goroutine 创建 tcp server，里面是个无限 for 循环
		exitFunc(protocol.TCPServer(n.tcpListener, n.tcpServer, n.logf))
	})
	// http
	if n.httpListener != nil {
		httpServer := newHTTPServer(n, false, n.getOpts().TLSRequired == TLSRequired)
		n.waitGroup.Wrap(func() {
			exitFunc(http_api.Serve(n.httpListener, httpServer, "HTTP", n.logf))
		})
	}
	// https
	if n.httpsListener != nil {
		httpsServer := newHTTPServer(n, true, true)
		n.waitGroup.Wrap(func() {
			exitFunc(http_api.Serve(n.httpsListener, httpsServer, "HTTPS", n.logf))
		})
	}
	
	// 省略部分代码 
	
	// 会一直阻塞
	err := <-exitCh
	return err
}

```

#### 创建 TCPServer
```go
protocol.TCPServer(n.tcpListener, n.tcpServer, n.logf)
```

进入 protocol 包 中的 tcp_server.go
```go
func TCPServer(listener net.Listener, handler TCPHandler, logf lg.AppLogFunc) error {
	logf(lg.INFO, "TCP: listening on %s", listener.Addr())

	var wg sync.WaitGroup

	for {
		// 等待客户端的连接，会阻塞
		clientConn, err := listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				logf(lg.WARN, "temporary Accept() failure - %s", err)
				runtime.Gosched()
				continue
			}
			// theres no direct way to detect this error because it is not exposed
			if !strings.Contains(err.Error(), "use of closed network connection") {
				return fmt.Errorf("listener.Accept() error - %s", err)
			}
			break
		}

		// 连接已经建立
		wg.Add(1)
		// 开启一个 goroutine 处理这个连接的数据I/O
		go func() {
			handler.Handle(clientConn)
			wg.Done()
		}()
	}

	// wait to return until all handler goroutines complete
	wg.Wait()

	logf(lg.INFO, "TCP: closing %s", listener.Addr())

	return nil
}
```
在一个无限 for 循环中 tcpListener 调用 Accept 等待客户端的连接，对每个新建立的连接都会创建一个 goroutine 把 conn 交给 TCPHandler 去处理，
这里的 TCPHandler 就是我们前面提到的 NSQD 中的 tcpServer。
使用 sync.WaitGroup 等待所有 TCPHandler 的 goroutine 处理完成。

#### TCPHandler 处理 I/O
TCPHandler 主要负责
- 对每个新创建的连接验证协议版本号
- 根据协议版本号创建对应的 protocol.Protocol
- 对新建立的连接 conn 创建一个 client
- prot.IOLoop(client) 处理 I/O

具体代码如下：
```go
func (p *tcpServer) Handle(conn net.Conn) {
	p.nsqd.logf(LOG_INFO, "TCP: new client(%s)", conn.RemoteAddr())

	// The client should initialize itself by sending a 4 byte sequence indicating
	// the version of the protocol that it intends to communicate, this will allow us
	// to gracefully upgrade the protocol away from text/line oriented to whatever...
	buf := make([]byte, 4)
	// 会阻塞，直到读取够4字节（  V2）
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		p.nsqd.logf(LOG_ERROR, "failed to read protocol version - %s", err)
		conn.Close()
		return
	}
	protocolMagic := string(buf)

	p.nsqd.logf(LOG_INFO, "CLIENT(%s): desired protocol magic '%s'",
		conn.RemoteAddr(), protocolMagic)

	var prot protocol.Protocol
	switch protocolMagic {
	case "  V2":
		prot = &protocolV2{nsqd: p.nsqd}
	default:
		protocol.SendFramedResponse(conn, frameTypeError, []byte("E_BAD_PROTOCOL"))
		conn.Close()
		p.nsqd.logf(LOG_ERROR, "client(%s) bad protocol magic '%s'",
			conn.RemoteAddr(), protocolMagic)
		return
	}
	// 对新建立的连接conn 创建一个 client
	client := prot.NewClient(conn)
	p.conns.Store(conn.RemoteAddr(), client)

	err = prot.IOLoop(client)
	if err != nil {
		p.nsqd.logf(LOG_ERROR, "client(%s) - %s", conn.RemoteAddr(), err)
	}

	p.conns.Delete(conn.RemoteAddr())
	// 下面一行等价 conn.Close()
	client.Close()
}
```
客户端与 NSQD 建立 TCP 连接之后，客户端需要发送4字节的4字节（[space][space][V][2]）的协议版本号数据给服务端 NSQD，
io.ReadFull(conn, buf) 读取协议版本号，目前 NSQ 支持的协议版本号是  V2（前面有2个空格），
如果协议版本号不对，会通过 `protocol.SendFramedResponse(conn, frameTypeError, []byte("E_BAD_PROTOCOL"))` 发送错误信息（E_BAD_PROTOCOL） 给客户端。

`&protocolV2{nsqd: p.nsqd}` 创建 protocolV2，`client := prot.NewClient(conn)` 对于新建立的连接创建一个对应的 client，这个 client 负责后面客户端发送过来的 Command 解析，处理等。

参考资料
- [NSQ TCP Protocol Spec](https://nsq.io/clients/tcp_protocol_spec.html)