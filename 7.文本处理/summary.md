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

