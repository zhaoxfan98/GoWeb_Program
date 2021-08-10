# 如何设计一个Web框架

通过Go语言来实现一个完整的框架设计，这框架中主要内容有第一小节介绍的Web框架的结构规划，例如采用MVC模式来进行开发，程序的执行流程设计等内容；第二小节介绍框架的第一个功能：路由，如何让访问的URL映射到相应的处理逻辑；第三小节介绍处理逻辑，如何设计一个公共的controller，对象继承之后处理函数中如何处理response和request；第四小节介绍框架的一些辅助功能，例如日志处理、配置信息等；第五小节介绍如何基于Web框架实现一个博客，包括博文的发表、修改、删除、显示列表等操作。

通过这么一个完整的项目例子，我期望能够让读者了解如何开发Web应用，如何搭建自己的目录结构，如何实现路由，如何实现MVC模式等各方面的开发内容。在框架盛行的今天，MVC也不再是神话。经常听到很多程序员讨论哪个框架好，哪个框架不好， 其实框架只是工具，没有好与不好，只有适合与不适合，适合自己的就是最好的，所以教会大家自己动手写框架，那么不同的需求都可以用自己的思路去实现。

## 13.1 项目规划

![](./dataflow.png?raw=true)

1. main.go作为应用入口，初始化一些运行博客所需要的基本资源，配置信息，监听端口
2. 路由功能检查HTTP请求，根据URL以及method来确定谁（控制层）来处理请求的转发资源
3. 如果缓存文件存在，它将绕过通常的流程执行，被直接发送给浏览器
4. 安全监测：应用程序控制器调用之前，HTTP请求和任一用户提交的数据将被过滤
5. 控制器装载模型、核心库、辅助函数，以及任何处理特定请求所需的其他资源，控制器主要负责处理业务逻辑
6. 输出视图层中渲染好的即将发送到Web浏览器中的内容。如果开启缓存，视图首先被缓存，将用于以后的常规请求。

### 目录结构

```
|——main.go         入口文件
|——conf            配置文件和处理模块
|——controllers     控制器入口
|——models          数据库处理模块
|——utils           辅助函数库
|——static          静态文件目录
|——views           视图库
```

### 框架设计

为了实现博客的快速搭建，打算基于上面的流程设计开发一个最小化的框架，框架包括路由功能、支持REST的控制器、自动化的模板渲染，日志系统、配置管理等。

## 13.2 自定义路由器设计

### HTTP路由

HTTP路由组件负责将HTTP请求交到对应的函数处理（或者是一个Struct的方法），如前面小节所描述的结构图，路由在框架中相当于一个事件处理器，而这个事件包括：

- 用户请求的路径，查询串信息
- HTTP的请求方法

路由器就是根据用户请求的事件信息转发到相应的处理函数（控制层）

### 默认的路由器实现

```
func fooHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

http.HandleFunc("/foo", fooHandler)

http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

路由的思想主要集中在两点

- 添加路由信息
- 根据用户请求转发到要执行的函数

Go默认的路由添加是通过函数http.Handle和http.HandleFunc等来添加，底层都是调用了`DefaultServeMux.Handle(pattern string, handler Handler)`，这个函数会把路由信息存储在一个map信息中`map[string]muxEntry`，这就解决了上面说的第一点

Go监听端口，然后接收到tcp连接会扔给`Handler`来处理，上面的例子默认nil即为`http.DefaultServeMux`，通过`DefaultServeMux.ServeHTTP`函数来进行调度，遍历之前存储的map路由信息，和用户访问的URL进行匹配，以查询对应注册的处理函数，这样就实现了上面所说的第二点。

```
for k, v := range mux.m {
	if !pathMatch(k, path) {
		continue
	}
	if h == nil || len(k) > n {
		n = len(k)
		h = v.h
	}
}
```

### beego框架路由实现

目前几乎所有的Web应用路由实现都是基于http默认的路由器，但是Go自带的路由器有几个限制：

- 不支持参数设定，例如/usr/:uid这种泛类型匹配
- 无法很好支持REST模式，无法限制访问的方法，例如上面的例子中，用户访问/foo，可以用GET、POST、DELETE、HEAD等方式访问
- 一般网站的路由规则太多了，编写繁琐。

beego框架的路由器基于上面的几点限制考虑设计了一种REST方式的路由实现，路由设计也是基于上面Go默认设计的两点来考虑：存储路由和转发路由

### 存储路由

针对前面所说的限制点，我们首先要解决参数支持就需要用到正则，第二和第三点我们通过一种变通的方法来解决，REST的方法对应到Struct的方法中去，然后路由到Struct而不是函数，这样在转发路由的时候就可以根据method来执行不同的方法

根据上面的思路，我们设计了两个数据类型controllerInfo(保存路径和对应的struct，这里是一个reflect.Type类型)和ControllerRegistor(routers是一个slice用来保存用户添加的路由信息，以及beego框架的应用信息)

```
type controllerInfo struct {
	regex          *regexp.Regexp
	params         map[int]string
	controllerType reflect.Type
}

type ControllerRegistor struct {
	routers     []*controllerInfo
	Application *App
}
```

ControllerRegistor对外的接口函数有

```
func (p *ControllerRegistor) Add(pattern string, c ControllerInterface)
```

### 静态路由实现

上面我们实现的动态路由的实现，Go的http包默认支持静态文件处理FileServer，由于我们实现了自定义的路由器，那么静态文件也需要自己设定，beego的静态文件夹路径保存在全局变量StaticDir中，StaticDir是一个map类型，实现如下：

```
func (app *App) SetStaticPath(url string, path string) *App {
	StaticDir[url] = path
	return app
}
```

### 转发路由

转发路由是基于ControllerRegistor里的路由信息来进行转发的

## 13.3 controller设计

传统的MVC框架大多是基于Action设计的后缀式映射，然而，现在Web流行REST风格的架构。尽管使用Filter或者rewriter能够通过URL重写实现REST风格的URL，但是为什么不直接设计一个全新的REST风格的 MVC框架呢？

### controller作用

MVC设计模式是目前Web应用开发中最常见的架构模式，通过分离Model View Controller，可以容易实现易于扩展的用户界面。Model指后台返回的数据；View指需要渲染的页面，通常是模板页面，渲染后的内容通常是HTML；Controller指Web开发人员编写的处理不同URL的控制器，如前面小节讲述的路由就是URL请求转发到控制器的过程，controller在整个的MVC框架中起到了一个核心的作用，负责处理业务逻辑，因此控制器是整个框架中必不可少的一部分，Model和View对于有些业务需求是可以不写的，例如没有数据处理的逻辑处理，没有页面输出的302调整之类的就不需要Model和View，但是controller这一环节是必不可少的。

### beego的REST设计

前面小节介绍了路由实现了注册struct的功能，而struct中实现了REST方式，因此我们需要设计一个用于逻辑处理controller的基类，这里主要设计了两个类型，一个struct、一个interface

```
type Controller struct {
	Ct        *Context
	Tpl       *template.Template
	Data      map[interface{}]interface{}
	ChildName string
	TplNames  string
	Layout    []string
	TplExt    string
}

type ControllerInterface interface {
	Init(ct *Context, cn string)    //初始化上下文和子类名称
	Prepare()                       //开始执行之前的一些处理
	Get()                           //method=GET的处理
	Post()                          //method=POST的处理
	Delete()                        //method=DELETE的处理
	Put()                           //method=PUT的处理
	Head()                          //method=HEAD的处理
	Patch()                         //method=PATCH的处理
	Options()                       //method=OPTIONS的处理
	Finish()                        //执行完成之后的处理		
	Render() error                  //执行完method对应的方法之后渲染页面
}
```

那么前面介绍的路由add函数的时候是定义了ControllerInterface类型，因此，只要我们实现这个接口就可以，所以我们的基类Controller实现如下的方法：

```
func (c *Controller) Init(ct *Context, cn string) {
	c.Data = make(map[interface{}]interface{})
	c.Layout = make([]string, 0)
	c.TplNames = ""
	c.ChildName = cn
	c.Ct = ct
	c.TplExt = "tpl"
}

func (c *Controller) Prepare() {

}

func (c *Controller) Finish() {

}

func (c *Controller) Get() {
	http.Error(c.Ct.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Post() {
	http.Error(c.Ct.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Delete() {
	http.Error(c.Ct.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Put() {
	http.Error(c.Ct.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Head() {
	http.Error(c.Ct.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Patch() {
	http.Error(c.Ct.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Options() {
	http.Error(c.Ct.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Render() error {
	if len(c.Layout) > 0 {
		var filenames []string
		for _, file := range c.Layout {
			filenames = append(filenames, path.Join(ViewsPath, file))
		}
		t, err := template.ParseFiles(filenames...)
		if err != nil {
			Trace("template ParseFiles err:", err)
		}
		err = t.ExecuteTemplate(c.Ct.ResponseWriter, c.TplNames, c.Data)
		if err != nil {
			Trace("template Execute err:", err)
		}
	} else {
		if c.TplNames == "" {
			c.TplNames = c.ChildName + "/" + c.Ct.Request.Method + "." + c.TplExt
		}
		t, err := template.ParseFiles(path.Join(ViewsPath, c.TplNames))
		if err != nil {
			Trace("template ParseFiles err:", err)
		}
		err = t.Execute(c.Ct.ResponseWriter, c.Data)
		if err != nil {
			Trace("template Execute err:", err)
		}
	}
	return nil
}

func (c *Controller) Redirect(url string, code int) {
	c.Ct.Redirect(code, url)
}
```

上面的controller基类已经实现了接口定义的函数，通过路由根据URL执行相应的controller的原则，会依次执行如下

```
Init()  初始化
Prepare()   执行之前的初始化，每个继承的子类可以来实现该函数
method()    根据不同的method执行不同的函数：GET、POST、PUT、HEAD等，子类来实现这些函数，如果没实现，那么默认都是403
Render()    可选，根据全局变量AutoRender来判断是否执行
Finish()    执行完之后执行的操作，每个继承的子类可以来实现该函数
```

