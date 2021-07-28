# 访问数据库

5.1小结介绍Go设计的一些驱动，介绍Go是如何设计数据库驱动接口的。5.2-5.4小节介绍目前使用的比较多的一些关系型数据驱动以及如何使用

目前NOSQL已经成为Web开发的一个潮流，很多应用采用了NOSQL作为数据库，而不是以前的缓存，5.6小节将介绍MongoDB和Redis两种NOSQL数据库。

## 5.1 database/sql接口

Go官方没有提供数据库驱动，而是为开发数据库驱动定义了一些标准接口，开发者可以根据定义的接口来开发相应的数据库驱动，这样的好处是只要是按照标准接口开发的代码，以后需要迁移数据库时，不需要任何修改。

### sql.Register

这个函数是用来注册数据库驱动的，当第三方开发者开发数据库驱动时，都会实现init函数，在init里面会调用这个Register(name string, driver driver.Driver)完成本驱动的注册

驱动调用示例

```
//https://github.com/mattn/go-sqlite3驱动
func init() {
	sql.Register("sqlite3", &SQLiteDriver{})
}
//https://github.com/mikespook/mymysql驱动
// Driver automatically registered in database/sql
var d = Driver{proto: "tcp", raddr: "127.0.0.1:3306"}
func init() {
	Register("SET NAMES utf8")
	sql.Register("mymysql", &d)
}
```

第三方数据库驱动都是通过调用这个函数来注册自己的数据库驱动名称以及相应的driver实现。在database/sql内部通过一个map来存储用户定义的相应驱动。

```
var drivers = make(map[string]driver.Driver)
drivers[name] = driver
```

因此通过database/sql的注册函数可以同时注册多个数据库驱动，只要不重复

```
  import (
  	"database/sql"
   	_ "github.com/mattn/go-sqlite3"
  )

```

> _的意思是引入后面的包而不直接使用这个包中定义的函数，变量等资源。包在引入的时候会自动调用包的init函数以完成对包的初始化。因此，引入上面的数据库驱动包之后会自动去调用init函数，然后在init函数里面注册这个数据库驱动，这样就可以在接下来的代码中直接使用这个数据库的驱动。

### driver.Driver

Driver是一个数据库驱动的接口，它定义了Open(name string)，这个方法返回一个数据库的Conn接口

```
type Driver interface {
    Open(name string) (Conn, error)
}
```

返回的Conn只能用来进行一次goroutine的操作，也就是说不能把这个Conn应用于Go的多个goroutine里面。

第三方驱动都会定义这个函数，它会解析name参数来获取相关数据库的连接信息，解析完成后，它将使用此信息来初始化一个Conn并返回

### driver.Conn

Conn是一个数据库连接的接口定义，它定义了一系列方法，这个Conn只能应用在一个goroutine里面，不能使用在多个goroutine里面（否则Go不知道某个操作究竟是由哪个goroutine发起的，从而导致数据混乱）

```
type Conn interface {
    Prepare(query string)(Stmt, error)
    Close() error
    Begin() (Tx, error)
}
```

Prepare函数返回与当前连接相关的执行sql语句的准备状态，可以进行查询、删除等操作

Close函数关闭当前的连接，执行释放连接拥有的资源的等清理工作。

Begin函数返回一个代表事务处理的Tx，通过它可以进行查询、更新等操作，或者对事务进行回滚、递交

### driver.Stmt

Stmt是一种准备好的状态，和Conn相关联，而且只能应用于一个goroutine中，不能应用于多个goroutine

```
type Stmt interface {
    Close() error
    NumInput() int
    Exec(args []value) (Result, error)
    Query(args []value) (Rows, error)
}
```

Close函数关闭当前的链接状态，但是如果当前正在执行query,query还是有效返回rows数据

NumInput函数返回当前预留参数的个数，当返回>=0时数据库驱动就会智能检查调用者的参数。当数据库驱动包不知道预留参数的时候，返回-1

Exec函数执行Prepare准备好的sql，传入参数执行update/insert等操作，返回Result数据

Query函数执行Prepare准备好的sql,传入需要的参数执行select操作，返回Rows结果集

### driver.Tx

事务处理一般就两个过程，递交或者回滚。数据库驱动里面也只需要实现这两个函数就可以

```
type Tx interface {
    Commit() error
    Rollback() error
}
```

### driver.Execer

这是一个Conn可选择实现的接口

```
type Excer interface {
    Exec(query string, args []Value) (Result, error)
}
```

如果这个接口没有定义，那么在调用DB.Exec，就会首先调用Prepare返回Stmt，然后执行Stmt的Exec，然后关闭Stmt

### driver.Result

这个是执行Update/Insert等操作返回的结果接口定义

```
type Result interface {
    LastInsertId() (int64, error)
    RowsAffected() (int64, error)
}
```

LastInsertId函数返回由数据库执行插入操作得到的自增ID号

RowsAffected函数返回执行Update/Insert等操作影响的数据条目数

### driver.Rows

Rows是执行查询返回的结果集接口定义

```
type Rows interface {
    Columns() []string
    Close() error
    Next(dest []Value) error
}
```

Columns函数返回查询数据库表的字段信息，这个返回的slice和sql查询的字段一一对应，而不是返回整个表的所有字段

Close函数用来关闭Rows迭代器。

Next函数用来返回下一条数据，把数据赋值给dest。dest里面的元素必须是driver.Value的值除了string，返回的数据里面所有的string都必须要转换成[]byte。如果最后没数据了，Next函数最后返回io.EOF。

### driver.RowsAffected

RowsAffected其实就是一个int64的别名，但是他实现了Result接口，用来底层实现Result的表示方式

```
type RowsAffected int64

func (RowsAffected) LastInsertId() (int64, error)

func (v RowsAffected) RowsAffected() (int64, error)
```

### driver.Value

Value其实就是一个空接口，他可以容纳任何的数据

```
type Value interface{}
```

drive的Value是驱动必须能够操作的Value，Value要么是nil，要么是下面的任意一种

```
int64
float64
bool
[]byte
string   [*]除了Rows.Next返回的不能是string.
time.Time
```

### driver.ValueConverter

ValueConverter接口定义了如何把一个普通的值转化成driver.Value的接口

```
type ValueConverter interface {
	ConvertValue(v interface{}) (Value, error)
}
```

在开发数据库驱动包里面实现这个接口的函数在很多地方会使用到，这个ValueConverter有很多好处：

- 转化driver.value到数据库表相应的字段，例如int64的数据如何转化成数据库表uint16字段
- 把数据库查询结果转化成driver.Value值
- 在scan函数里面如何把driver.Value值转化成用户定义的值

### driver.Valuer

Valuer接口定义了返回一个driver.Value的方式

```
type Valuer interface {
	Value() (Value, error)
}
```

很多类型都实现了这个Value方法，用来自身与driver.Value的转化

通过以上讲解，驱动开发有了一个基本的了解。一个驱动只要实现了这些接口就能完成增删改查等基本操作，剩下的就是与相应的数据库进行数据交互等细节问题了，在此不再赘述

### database/sql

database/sql在database/sql/driver提供的接口基础上定义了一些更高阶的方法，用以简化数据库操作，同时内部还建议性地实现了一个Conn pool

```
type DB struct {
	driver 	 driver.Driver
	dsn    	 string
	mu       sync.Mutex // protects freeConn and closed
	freeConn []driver.Conn
	closed   bool
}
```

我们可以看到Open函数返回的是DB对象，里面有一个freeConn，它就是那个简易的连接池。它的实现相当简单或者说简陋，就是当执行db.prepare -> db.prepareDC的时候会defer dc.releaseConn，然后调用db.putConn，也就是把这个连接放入连接池，每次调用db.conn的时候会先判断freeConn的长度是否大于0，大于0说明有可以复用的conn，直接拿出来用就是了，如果不大于0，则创建一个conn，然后再返回之。

## 5.2 使用MySQL数据库

- https://github.com/go-sql-driver/mysql 支持database/sql，全部采用go写。

接下来的例子使用以上驱动，用的理由是

- 这个驱动比较新，维护比较好
- 完全支持database/sql接口
- 支持keeplive，保持长连接

sql.Open()函数用来打开一个注册过的数据库驱动，go-sql-driver中注册了mysql这个数据库驱动，第二个参数是DSN(Data Source Name),它是go-sql-driver定义的一些数据库链接和配置信息。

db.Prepare()函数用来返回准备要执行的sql操作，然后返回准备完毕的执行状态

db.Query()函数用来直接执行Sql返回Rows结果

stmt.Exec()函数用来执行stmt准备好的SQL语句

## 5.3 使用SQLite数据库

SQLite是一个开源的嵌入式关系数据库，实现自包容、零配置、支持事务的SQL数据库引擎。特点是高度便携、使用方便、结构紧凑、高效、可靠。与其他数据库管理系统不同，SQLite 的安装和运行非常简单，在大多数情况下,只要确保SQLite的二进制文件存在即可开始创建、连接和使用数据库。如果您正在寻找一个嵌入式数据库项目或解决方案，SQLite是绝对值得考虑。SQLite可以说是开源的Access。

## 5.4 使用PostgreSQL数据库

PostgreSQL 是一个自由的对象-关系数据库服务器(数据库管理系统)，它在灵活的 BSD-风格许可证下发行。它提供了相对其他开放源代码数据库系统(比如 MySQL 和 Firebird)，和对专有系统比如 Oracle、Sybase、IBM 的 DB2 和 Microsoft SQL Server的一种选择。

PostgreSQL和MySQL比较，它更加庞大一点，因为它是用来替代Oracle而设计的。所以在企业应用中采用PostgreSQL是一个明智的选择。

MySQL被Oracle收购之后正在逐步的封闭（自MySQL 5.5.31以后的所有版本将不再遵循GPL协议），鉴于此，将来我们也许会选择PostgreSQL而不是MySQL作为项目的后端数据库。

## 5.5 使用Beego orm库进行ORM开发

beego orm是一个Go进行ORM操作的库，采用了Go style方式对数据库进行操作，实现了struct到数据表记录的映射。beego orm是一个十分轻量的Go ORM框架。

beego orm是支持database/sql标准接口的ORM库，所以理论上来说，只要数据库驱动支持database/sql接口就可以无缝的接入beego orm

**安装

```
go get github.com/astaxie/beego
```

**如何初始化

```
import (
	"database/sql"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	//注册驱动
	orm.RegisterDriver("mysql", orm.DRMySQL)
	//设置默认数据库
	orm.RegisterDataBase("default", "mysql", "root:root@/my_db?charset=utf8", 30)
	//注册定义的model
    	orm.RegisterModel(new(User))

   	// 创建table
        orm.RunSyncdb("default", false, true)
}
```

MySQL配置：

```
//导入驱动
//_ "github.com/go-sql-driver/mysql"

//注册驱动
orm.RegisterDriver("mysql", orm.DR_MySQL)

// 设置默认数据库
//mysql用户：root ，密码：zxxx ， 数据库名称：test ， 数据库别名：default
 orm.RegisterDataBase("default", "mysql", "root:zxxx@/test?charset=utf8")
```

导入必须的package之后，需要打开数据库的链接，然后创建一个beego orm对象，如下所示

```
func main() {
    o := orm.NewOrm()
}
```

## 5.6 NOSQL数据库操作

NoSQL指的是非关系型的数据库。随着Web2.0的兴起，传统的关系数据库在应付Web2.0网站，特别是超大规模和高并发的SNS类型的Web2.0纯动态网站已经显得力不从心，暴露了很多难以克服的问题，而非关系型的数据库则由于其本身的特点得到了非常迅速的发展。

目前流行的NOSQL主要有redis、mongoDB、Cassandra和Membase等。这些数据库都有高性能、高并发读写等特点

### redis



### mongoDB

