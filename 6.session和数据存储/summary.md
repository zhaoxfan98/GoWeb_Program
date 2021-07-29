# session和数据存储

Web开发中一个很重要的议题就是如何做好用户的整个浏览过程的控制，因为HTTP协议是无状态的，所以用户的每一次请求都是无状态的，我们不知道在整个Web操作过程中哪些连接与该用户有关，我们应该如何来解决这个问题呢？

Web里面经典的解决方案是cookie和session，Cookie机制是一种客户端机制，把用户数据保存在客户端，而session机制是一种服务器端的机制，服务器使用一种类似于散列表的结构来保存信息，每一个网站访客都会被分配给一个唯一的标志符，即SessionID，它的存放形式无非两种：要么经过URL传递，要么保存在客户端的cookies里，当然，你也可以将session保存到数据库里，这样会更安全，但效率方面会有所下降。

6.1小节里面讲介绍session机制和cookie机制的关系和区别，6.2讲解Go语言如何来实现session，里面讲实现一个简易的session管理器，6.3小节讲解如何防止session被劫持的情况，如何有效的保护session。我们知道session其实可以存储在任何地方，6.4小节里面实现的session是存储在内存中的，但是如果我们的应用进一步扩展了，要实现应用的session共享，那么我们可以把session存储在数据库中(memcache或者redis)，6.5小节将详细的讲解如何实现这些功能。

## 6.1 session和cookie

cookie简而言之就是在本地计算机保存一些用户操作的历史信息，并在用户再次访问该站点时浏览器通过HTTP协议将本地Cookie内容发送给服务器，从而完成验证，或继续上一步操作。

![cookie的原理图](./cookie.png)

session就是在服务器上保存用户操作的历史信息。服务器使用session id来标识session，session id由服务器负责产生，保证随机性与唯一性，相当于一个随机密钥，避免在握手或传输中暴露用户真实密码。但该方式下，仍然需要将发送请求的客户端与session进行对应，所以可以借助cookie机制来获取客户端的标识（即session id），也可以通过GET方式将id提交给服务器。

![session的原理图](./session.png)

**cookie**

Cookie是由浏览器维持的，存储在客户端的一小段文本信息，伴随着用户请求和页面在web服务器和浏览器之间传递。用户每次访问站点时，Web应用程序都可以读取cookie包含的信息。

如果不设置过期时间，则表示这个cookie的生命周期为从创建到浏览器关闭为止，只要关闭浏览器窗口，cookie就消失了。这种生命期为浏览会话期的cookie被称为会话cookie。会话cookie一般不保存在硬盘上而是保存在内存里。

如果设置了过期时间，浏览器就会把cookie保存到硬盘上，关闭后再次打开浏览器，这些cookie依然有效直到超过设定的过期时间。

### Go设置cookie

Go通过net/http包中的SetCookie来设置：

```
http.SetCookie(w ResponseWriter, cookie *Cookie)
```

w表示需要写入的response，cookie是一个struct，让我们来看一下cookie对象是怎么样的

```
type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string

// MaxAge=0 means no 'Max-Age' attribute specified.
// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}
```

我们来看一个例子，如何设置cookie

```
expiration := time.Now()
expiration = expiration.AddDate(1, 0, 0)
cookie := http.Cookie{Name: "username", Value: "astaxie", Expires: expiration}
http.SetCookie(w, &cookie)
```

### Go读取cookie

```
cookie, _ := r.Cookie("username")
fmt.Fprint(w, cookie)

//another way
for _, cookie := range r.Cookies() {
	fmt.Fprint(w, cookie.Name)
}
```

**session**

含义是指有始有终的一系列动作/消息。session在Web开发环境下的语义又有了扩展，它的含义是指一类用来在客户端与服务器端之间保持状态的解决方案。有时候Session也用来指这种解决方案的存储结构。

session机制是一种服务器端的机制，服务器使用一种类似于散列表的结构来保存信息

### 小结

如上文所述，session和cookie的目的相同，都是为了克服http协议无状态的缺陷，但完成的方法不同。session通过cookie，在客户端保存session id，而将用户的其他会话消息保存在服务端的session对象中，与此相对的，cookie需要将所有信息都保存在客户端。因此cookie存在着一定的安全隐患，例如本地cookie中保存的用户名密码被破译，或cookie被其他网站收集（例如：1. appA主动设置域B cookie，让域B cookie获取；2. XSS，在appA上通过javascript获取document.cookie，并传递给自己的appB）。

通过上面的一些简单介绍我们了解了cookie和session的一些基础知识，知道他们之间的联系和区别，做web开发之前，有必要将一些必要知识了解清楚，才不会在用到时捉襟见肘，或是在调bug时如无头苍蝇乱转。接下来的几小节我们将详细介绍session相关的知识。

## 6.2 Go如何使用session

