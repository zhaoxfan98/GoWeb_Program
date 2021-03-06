# 4.表单

## 4.1 处理表单的输入

获取请求方法是通过r.Method来完成，这是个字符串类型的变量，返回GET、POST、PUT等method信息

Request本身也提供了FormValue()函数来获取用户提交的参数。如r.Form["username"]也可写成r.FormValue("username")。调用r.FormValue时会自动调用r.ParseForm，所以不必提前调用。r.FormValue只会返回同名参数的第一个，若参数不存在则返回空字符串。

## 4.2 验证表单的输入

平时编写Web应用主要有两方面的数据验证，一个是在页面端的js验证（目前在这方面有很多的插件库，比如ValidationJS插件），一个是在服务器端的验证，这节讲解如何在服务器端验证。

**数字**

先转化成int类型， 然后进行处理

```
getint,err:=strconv.Atoi(r.Form.Get("age"))
if err!=nil{
	//数字转化出错了，那么可能就不是数字
}

//接下来就可以判断这个数字的大小范围了
if getint >100 {
	//太大了
}
```

*我们应该尽量避免使用正则表达式，因为速度会比较慢。但是在目前机器性能那么强劲的情况下，对于这种简单的正则表达式效率和类型转换函数是没有什么差别的。*

```

//验证15位身份证，15位的是全部数字
if m, _ := regexp.MatchString(`^(\d{15})$`, r.Form.Get("usercard")); !m {
	return false
}

//验证18位身份证，18位前17位为数字，最后一位是校验位，可能为数字或字符X。
if m, _ := regexp.MatchString(`^(\d{17})([0-9]|X)$`, r.Form.Get("usercard")); !m {
	return false
}

```

上述服务器端表单元素验证

## 4.3 预防跨站脚本

现在的网站包含大量的动态内容以提高用户体验，比过去要复杂得多。所谓动态内容，就是根据用户环境和需要，Web应用程序能够输出相应的内容。动态站点会受到一种名为“跨站脚本攻击”（Cross Site Scripting, 安全专家们通常将其缩写成 XSS）的威胁，而静态站点则完全不受其影响。

攻击者通常会在有漏洞的程序中插入JavaScript、VBScript、 ActiveX或Flash以欺骗用户。一旦得手，他们可以盗取用户帐户信息，修改用户设置，盗取/污染cookie和植入恶意广告等。

对XSS最佳的防护应该结合以下两种方法：一是验证所有输入数据，有效检测攻击(这个我们前面小节已经有过介绍);另一个是对所有输出数据进行适当的处理，以防止任何已成功注入的脚本在浏览器端运行。

- func HTMLEscape(w io.Writer, b []byte) //把b进行转义之后写到w
- func HTMLEscapeString(s string) string //转义s之后返回结果字符串
- func HTMLEscaper(args ...interface{}) string //支持多个参数一起转义，返回结果字符串

## 4.4 防止多次提交表单

解决方案就是在表单中添加一个带有唯一值的隐藏字段。在验证表单时，先检查带有该唯一值的表单是否已经递交过了。如果是，拒绝再次递交；如果不是，则处理表单进行逻辑处理。另外，如果采用了Ajax模式递交表单的话，当表单递交后，通过js来禁用表单的递交按钮

## 4.5 处理文件上传

要使表单能够上传文件，首先第一步就是添加form的enctype属性，enctype属性有如下三种情况：

1. application/x-www-form-urlencoded   表示在发送前编码所有字符（默认）
2. multipart/form-data	  不对字符编码。在使用包含文件上传控件的表单时，必须使用该值。
3. text/plain	  空格转换为 "+" 加号，但不对特殊字符编码。

上传文件主要三步处理：

1. 表单中增加enctype="multipart/form-data"
2. 服务端调用r.ParseMultipartForm，把上传的文件存储在内存和临时文件中
3. 使用r.FormFile获取文件句柄，然后对文件进行存储等处理

客户端通过multipart.Write把文件的文本流写入一个缓存中，然后调用http的Post方法把缓存传到服务器

## 小结

学习Go如何处理表单信息，通过用户登录、文件上传的例子展示

能够了解客户端和服务端是如何进行数据上的交互，客户端将数据传递给服务器系统，服务器接受数据又把处理结果反馈给客户端。