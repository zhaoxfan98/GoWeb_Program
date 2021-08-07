# 错误处理，调试和测试

11.1小节将介绍Go语言中如何处理错误，如何设计自己的包、函数的错误处理，11.2小节将介绍如何使用GDB来调试我们的程序，动态运行情况下各种变量信息，运行情况的监控和调试。

11.3小节将对Go语言中的单元测试进行深入的探讨，并示例如何来编写单元测试，Go的单元测试规则规范如何定义，以保证以后升级修改运行相应的测试代码就可以进行最小化的测试。

长期以来，培养良好的调试、测试习惯一直是很多程序员逃避的事情，所以现在你不要再逃避了，就从你现在的项目开发，从学习Go Web开发开始养成良好的习惯。

## 11.1 错误处理

Go语言主要的设计准则是：简洁、明白，简洁是指语法和C类似，相当的简单，明白是指任何语句都是很明显的，不含有任何隐含的东西，在错误处理方案的设计中也贯彻了这一思想。我们知道在C语言里面是通过返回-1或者NULL之类的信息来表示错误，但是对于使用者来说，不查看相应的API说明文档，根本搞不清楚这个返回值究竟代表什么意思，比如:返回0是成功，还是失败,而Go定义了一个叫做error的类型，来显式表达错误。在使用时，通过把返回的error变量与nil的比较，来判定操作是否成功。

类似于os.Open函数，标准包中所有可能出错的API都会返回一个error变量，以方便错误处理，这个小节将详细地介绍error类型的设计，和讨论开发Web应用中如何更好地处理error。

### ERROR类型

error类型是一个接口类型，这是它的定义
```
type error interface {
    Error() string
}
```

error是一个内置的接口类型，可以在/builtin/包下面找到相应的定义。而我们在很多内部包里面用到的 error是errors包下面的实现的私有结构errorString

```
// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
```

可以通过errors.New把一个字符串转化为errorString，以得到一个满足接口error的对象，其内部实现如下：

```
// New returns an error that formats as the given text.
func New(text string) error {
	return &errorString{text}
}
```

下面这个例子演示了如何使用errors.New:
```
func Sqrt(f float64) (float64, error) {
	if f < 0 {
		return 0, errors.New("math: square root of negative number")
	}
	// implementation
}
```

### 自定义Error

通过上面的介绍我们知道error是一个interface，所以在实现自己的包的时候，通过定义实现此接口的结构，我们就可以实现自己的错误定义，请看来自Json包的示例：

```
type SyntaxError struct {
	msg    string // 错误描述
	Offset int64  // 错误发生的位置
}
func (e *SyntaxError) Error() string { return e.msg }
```

Offset字段在调用Error的时候不会被打印，但是我们可以通过类型断言获取错误类型，然后可以打印相应的错误信息，请看下面的例子:

```
if err := dec.Decode(&val); err != nil {
	if serr, ok := err.(*json.SyntaxError); ok {
		line, col := findLine(f, serr.Offset)
		return fmt.Errorf("%s:%d:%d: %v", f.Name(), line, col, err)
	}
	return err
}
```

> 需要注意的是，函数返回自定义错误时，返回值推荐设置为error类型，而非自定义错误类型，特别需要注意的是不应预声明自定义错误类型的变量。例如：

```
func Decode() *SyntaxError { // 错误，将可能导致上层调用者err!=nil的判断永远为true。
        var err *SyntaxError     // 预声明错误变量
        if 出错条件 {
            err = &SyntaxError{}
        }
        return err               // 错误，err永远等于非nil，导致上层调用者err!=nil的判断始终为true
    }
```

上面例子简单的演示了如何自定义Error类型。但是如果我们还需要更复杂的错误处理呢？此时，我们来参考一下net包采用的方法：

```
package net

type Error interface {
    error
    Timeout() bool   // Is the error a timeout?
    Temporary() bool // Is the error temporary?
}
```

在调用的地方，通过类型断言err是不是net.Error,来细化错误的处理，例如下面的例子，如果一个网络发生临时性错误，那么将会sleep 1秒之后重试：

```
if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
	time.Sleep(1e9)
	continue
}
if err != nil {
	log.Fatal(err)
}
```

### 错误处理

Go在错误处理上采用了与C类似的检查返回值的方式，而不是其他多数主流语言采用的异常方式，这造成了代码编写上的一个很大的缺点:错误处理代码的冗余，对于这种情况是我们通过复用检测函数来减少类似的代码。

### 总结

在程序设计中，容错是相当重要的一部分工作，在Go中它是通过错误处理来实现的，error虽然只是一个接口，但是其变化却可以有很多，我们可以根据自己的需求来实现不同的处理

## 11.2 使用GDB调试


## 11.3 Go怎么写测试用例

Go语言中自带有一个轻量级的测试框架testing和自带的go test命令来实现单元测试和性能测试，testing框架和其他语言中的测试框架类似，你可以基于这个框架写针对相应函数的测试用例，也可以基于该框架写相应的压力测试用例，那么接下来让我们一一来看一下怎么写。

- 文件名必须是_test.go结尾的，这样在执行go test的时候才会执行到相应的代码
- 必须import testing这个包
- 所有的测试用例函数必须是Test开头
- 测试用例会按照源代码中写的顺序依次执行
- 测试函数TestXxx()的参数是testing.T，我们可以使用该类型来记录错误或者是测试状态
- 测试格式：func TestXxx (t *testing.T),Xxx部分可以为任意的字母数字的组合，但是首字母不能是小写字母[a-z]，例如Testintdiv是错误的函数名。
- 函数中通过调用testing.T的Error, Errorf, FailNow, Fatal, FatalIf方法，说明测试不通过，调用Log方法用来记录测试的信息。

