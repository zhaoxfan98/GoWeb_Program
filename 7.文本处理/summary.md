# 文本处理

## 7.1 XML处理

可以通过xml包的Unmarshal函数来解析xml
```
func Unmarshal(data []byte, v interface{}) error
```

XML本质上是一种树形的数据格式，而我们可以定义与之匹配的go 语言的 struct类型，然后通过xml.Unmarshal来将xml中的数据解析成对应的struct对象。

Unmarshal函数定义了两个参数，第一个是XML数据流，第二个是存储的对应类型，目前支撑Struct slice String。XML包内部采用了反射来进行数据的映射，所以v里面的字段必须是导出的。

> 注意： 为了正确解析，go语言的xml包要求struct定义中的所有字段必须是可导出的（即首字母大写）

## 7.2 JSON处理

能够被赋值的字段必须是可导出字段（即首字母大写）。同时JSON解析的时候只会解析能找得到的字段，找不到的字段会被忽略，这样的一个好处是：当接收到一个很大的JSON数据结构而只想获取其中的部分数据的时候，只需将想要的数据对应的字段名大写，即可轻松解决这个问题

**解析JSON**

假设有如下的JSON数据
```
b := []byte(`{"Name":"Wednesday","Age":6,"Parents":["Gomez","Morticia"]}`)
```

如果在我们不知道他的结构的情况下，把它解析到interface{}里面

```
var f interface{}
err := json.Unmarshal(b, &f)
```

这个时候f里面存储了一个map类型，他们的key是string，值存储在空的interface{}里
```
f = map[string]interface{}{
	"Name": "Wednesday",
	"Age":  6,
	"Parents": []interface{}{
		"Gomez",
		"Morticia",
	},
}
```

那么如何来访问这些数据呢？通过断言的方式：
```
m := f.(map[string]interface{})
```
通过断言之后，你就可以通过如下方式来访问里面的数据了
```
for k, v := range m {
	switch vv := v.(type) {
	case string:
		fmt.Println(k, "is string", vv)
	case int:
		fmt.Println(k, "is int", vv)
	case float64:
		fmt.Println(k,"is float64",vv)
	case []interface{}:
		fmt.Println(k, "is an array:")
		for i, u := range vv {
			fmt.Println(i, u)
		}
	default:
		fmt.Println(k, "is of a type I don't know how to handle")
	}
}
```

通过上面的示例可以看到，通过interface{}与type assert的配合，我们就可以解析未知结构的JSON数了。


## 7.3 正则处理

## 7.4 模板处理

MVC设计模式中，Model处理数据，View展现结果，Controller控制用户的请求，至于VIew层的处理，在很多动态语言里面都是通过在静态HTML中插入动态语言生成的数据来实现的。例如JSP中通过插入<%= ... =%>

**Go模板的使用**

使用template包来进行模板处理，使用类似Parse、ParseFile、Execute等方法从文件或者字符串加载模板，然后执行模板的merge操作

Go语言的模板通过{{}}来包含需要在渲染时被替换的字段，{{.}}表示当前对象，这和Java或者C++中的this类似，如果要访问当前对象的字段通过{{.FileName}}，需要注意的是：这个字段必须是导出的（字段首字母必须是大写的），否则渲染的时候会报错

**输出嵌套字段内容**

如果字段里面还有对象，如何来循环的输出这些内容呢？
可以使用{{with ...}} {{end}} {{range ...}} {{end}}来进行数据的输出

通过以上了解如何把动态数据与模板融合、如何输出循环数据、如何自定义函数、如何嵌套模板等等。通过模板技术的应用，可以完成MVC模式中V的处理，接下来介绍处理M和C

## 7.5 文件操作

**目录操作**

- func Mkdir(name string, perm FileMode) error

创建名称为name的目录，权限设置是perm，例如0777

- func MkdirAll(path string, perm FileMode) error

根据path创建多级子目录，例如a/b/c

- func Remove(name string) error

删除name目录，当目录下有文件或其他目录时报错

- func RemoveAll(path string) error

根据path删除多级子目录，如果path是单个名称，那么该目录下的子目录全部删除

**文件操作**

- func Create(name string) (file *File, err Error)

根据提供的文件名创建新的文件，返回一个文件对象，默认权限是0666的文件，返回的文件对象是可读写的

- func NewFile(fd uintptr, name string) *File

根据文件描述符创建相应的文件，返回一个文件对象

- func Open(name string)(file *File, err Error)

该方法打开一个名称为name的文件，但是只读方式，内部实现其实调用了Openfile

- func OpenFile(name string, flag int, perm uint32)(file *File, err Error)

打开名称为name的文件，flag是打开的方式，只读、读写等，perm是权限

- func (file *File) Write(b []byte) (n int, err Error)

写入byte类型的信息到文件

- func (file *File) WriteAt(b []byte, off int64) (n int, err Error)

在指定位置开始写入byte类型的信息

- func (file *File) WriteString(s string) (ret int, err Error)

写入string信息到文件

- func (file *File) Read(b []byte) (n int, err Error)

读取数据到b中

- func (file *File) ReadAt(b []byte, off int64) (n int, err Error)

从off开始读取数据到b中

- func Remove(name string) Error

Go语言里面删除文件和删除文件夹是同一个函数, 调用该函数就可以删除文件名为name的文件

## 7.6 字符串操作

我们经常需要对字符串进行分割、连接、转换等操作，本节通过GO标准库中的strings和strconv两个包中的函数来讲解如何进行有效快速的操作

**字符串操作**

下面这些函数来自于strings包

- func Contains(s, substr string) bool

字符串s中是否包含substr，返回bool值

- func Join(a []string, sep string) string

字符串链接，把slice a通过sep链接起来

- func Index(s, sep string) int

在字符串s中查找sep所在的位置，返回位置值，找不到返回-1

- func Repeat(s string, count int) string

重复s字符串count次，最后返回重复的字符串

- func Replace(s, old, new string, n int) string

在s字符串中，把old字符串替换为new字符串，n表示替换的次数，小于0表示全部替换

- func Split(s, sep string) []string

把s字符串按照sep分割，返回slice

- func Trim(s string, cutset string) string

在s字符串的头部和尾部去除cutset指定的字符串

- func Fields(s string) []string

去除s字符串的空格符，并且按照空格分割返回slice

**字符串转换**

字符串转化的函数在strconv中

```
func main() {
	//Append 系列函数将整数等转换为字符串后，添加到现有的字节数组中。
	str := make([]byte, 0, 100)
	str = strconv.AppendInt(str, 4567, 10)
	str = strconv.AppendBool(str, false)
	str = strconv.AppendQuote(str, "abcdefg")
	str = strconv.AppendQuoteRune(str, '单')
	fmt.Println(string(str))

	//Format 系列函数把其他类型的转换为字符串
	a := strconv.FormatBool(false)
	b := strconv.FormatFloat(123.23, 'g', 12, 64)
	c := strconv.FormatInt(1234, 10)
	d := strconv.FormatUint(12345, 10)
	e := strconv.Itoa(1023)
	fmt.Println(a, b, c, d, e)

	//Parse 系列函数把字符串转换为其他类型
	a, err := strconv.ParseBool("false")
	checkError(err)
	b, err := strconv.ParseFloat("123.23", 64)
	checkError(err)
	c, err := strconv.ParseInt("1234", 10, 64)
	checkError(err)
	d, err := strconv.ParseUint("12345", 10, 64)
	checkError(err)
	e, err := strconv.Atoi("1023")
	checkError(err)
	fmt.Println(a, b, c, d, e)
}
```

## 7.7 小结

这一章给大家介绍了一些文本处理的工具，包括XML、JSON、正则和模板技术，XML和JSON是数据交互的工具，通过XML和JSON你可以表达各种含义，通过正则你可以处理文本(搜索、替换、截取)，通过模板技术你可以展现这些数据给用户。这些都是你开发Web应用过程中需要用到的技术，通过这个小节的介绍你能够了解如何处理文本、展现文本。