# Web服务

Web服务可以在HTTP协议的基础上通过XML或者JSON来交换信息。如果你想知道上海的天气预报、中国石油的股价或者淘宝商家的一个商品信息，你可以编写一段简短的代码，通过抓取这些信息然后通过标准的接口开放出来，就如同你调用一个本地函数并返回一个值。

目前主流的有如下几种Web服务：REST、SOAP。

REST请求是很直观的，因为REST是基于HTTP协议的一个补充，他的每一次请求都是一个HTTP请求，然后根据不同的method来处理不同的逻辑。

SOAP是W3C在跨网络信息传递和远程计算机函数调用方面的一个标准。但是SOAP非常复杂，其完整的规范篇幅很长，而且内容仍然在增加。

Go追求的是性能、简单。

## 8.1 Socket编程

**什么是Socket？**

Socket起源于Unix，而Unix基本哲学之一就是“一切皆文件”，都可以用“打开open->读写write/read->关闭close”模式来操作。Socket就是该模式的一个实现，网络的Socket数据传输是一种特殊的I/O，Socket也是一种文件描述符。Socket也具有一个类似于打开文件的函数调用：Socket()，该函数返回一个整形的Socket描述符，随后的连接建立，数据传输等操作都是通过该Socket实现的。

常用的Socket类型有两种：流式socket和数据报式Socket。流式是一种面向连接的socket，针对于面向连接的TCP服务应用；数据报式socket是一种无连接的socket，对应于无连接的UDP服务应用。

**socket如何通信**

如何通信首要解决的问题是如何唯一标识一个进程，否则通信无从谈起！本地可以通过进程PID来标识进程，但网络中行不通。但TCP/IP协议簇帮我们解决了这个问题，网络层的"ip地址"可以唯一标识网络中的主机，而传输层的"协议+端口"可以唯一标识主机中的应用程序（进程）。这样利用三元组（ip地址，协议，端口）就可以标识网络的进程了，网络中需要互相通信的进程，就可以利用这个标志在他们之间进行交互。

**Socket基础知识**

### TCP Socket

当我们知道如何通过网络端口访问一个服务时，那么我们能够做什么呢？*作为客户端来说，我们可以通过向远端某台机器的的某个网络端口发送一个请求，然后得到在机器的此端口上监听的服务反馈的信息。作为服务端，我们需要把服务绑定到某个指定端口，并且在此端口上监听，当有客户端来访问时能够读取信息并且写入反馈信息。*

在Go语言的net包中有一个类型TCPConn，这个类型可以用来作为客户端和服务器端交互的通道，他有两个主要的函数：

```
func (c *TCPConn) Write(b []byte) (int, error)
func (c *TCPConn) Read(b []byte) (int, error)
```

TCPConn可以用在客户端和服务端来读写数据。

还有我们需要知道一个TCPAddr类型，他表示一个TCP的地址信息，他的定义如下：

```
type TCPAddr struct {
	IP IP
	Port int
	Zone string // IPv6 scoped addressing zone
}
```

在Go语言中通过ResolveTCPAddr获取一个TCPAddr

```
func ResolveTCPAddr(net, addr string) (*TCPAddr, os.Error)
```

- net参数是"tcp4"、"tcp6"、"tcp"中的任意一个，分别表示TCP(IPv4-only), TCP(IPv6-only)或者TCP(IPv4, IPv6的任意一个)。
- addr表示域名或者IP地址，例如"www.google.com:80" 或者"127.0.0.1:22"。

### TCP client

Go语言中通过net包中的DialTCP函数来建立一个TCP连接，并返回一个TCPConn类型的对象，当连接建立时服务端也创建一个同类型的对象，此时客户端和服务器端通过各自拥有的TCPConn对象来进行数据交换。一般而言，客户端通过TCPConn对象将请求信息发送到服务器端，读取服务器端响应的信息。服务器端读取并解析来自客户端的请求，并返回应答信息，这个连接只有当任一端关闭了连接之后才失效，不然这连接可以一直在使用。建立连接的函数定义如下：

```
func DialTCP(network string, laddr, raddr *TCPAddr) (*TCPConn, error)
```

- network参数是"tcp4"、"tcp6"、"tcp"中的任意一个，分别表示TCP(IPv4-only)、TCP(IPv6-only)或者TCP(IPv4,IPv6的任意一个)
- laddr表示本机地址，一般设置为nil
- raddr表示远程的服务地址

### TCP server

也可以通过net包来创建一个服务器端程序，在服务器端我们需要绑定服务到指定的非激活端口，并监听此端口，当有客户端请求到达的时候可以接收到来自客户端连接的请求。net包中有相应功能的函数，函数定义如下：

```
func ListenTCP(network string, laddr *TCPAddr) (*TCPListener, error)
func (l *TCPListener) Accept() (Conn, error)
```


### 控制TCP连接

TCP有很多连接控制函数，我们平常用到比较多的有如下几个函数：

- func DialTimeout(net, addr string, timeout time.Duration) (Conn, error)

设置建立连接的超时时间，客户端和服务器端都适用，当超过设置时间时，连接自动关闭。

- func (c *TCPConn) SetReadDeadline(t time.Time) error
- func (c *TCPConn) SetWriteDeadline(t time.Time) error

用来设置写入/读取一个连接的超时时间。当超过设置时间时，连接自动关闭。

- func (c *TCPConn) SetKeepAlive(keepalive bool) os.Error

设置keepAlive属性。操作系统层在tcp上没有数据和ACK的时候，会间隔性的发送keepalive包，操作系统可以通过该包来判断一个tcp连接是否已经断开，在windows上默认2个小时没有收到数据和keepalive包的时候认为tcp连接已经断开，这个功能和我们通常在应用层加的心跳包的功能类似。

### UDP Socket

Go语言包中处理UDP Socket和TCP Socket不同的地方就是在服务端处理多个客户端请求数据包的方式不同，UDP缺失了对客户端连接请求的Accept函数。其他基本几乎一模一样，只有TCP换成了UDP而已。UDP的几个主要函数如下所示

```
func ResolveUDPAddr(net, addr string) (*UDPAddr, os.Error)
func DialUDP(net string, laddr, raddr *UDPAddr) (c *UDPConn, err os.Error)
func ListenUDP(net string, laddr *UDPAddr) (c *UDPConn, err os.Error)
func (c *UDPConn) ReadFromUDP(b []byte) (n int, addr *UDPAddr, err os.Error)
func (c *UDPConn) WriteToUDP(b []byte, addr *UDPAddr) (n int, err os.Error)
```

## 8.2 WebSocket

