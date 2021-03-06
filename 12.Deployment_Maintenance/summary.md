# 部署与维护

本章我们将通过四个小节来介绍这些小细节的处理，第一小节介绍如何在生产服务上记录程序产生的日志，如何记录日志，第二小节介绍发生错误时我们的程序如何处理，如何保证尽量少的影响到用户的访问，第三小节介绍如何来部署Go的独立程序，由于目前Go程序还无法像C那样写成daemon，那么我们如何管理这样的进程程序后台运行呢？第四小节将介绍应用数据的备份和恢复，尽量保证应用在崩溃的情况能够保持数据的完整性。

## 12.1 应用日志

Go语言中提供了一个简易的log包，我们使用该包可以方便的实现日志记录的功能，这些日志都是基于fmt包的打印再结合panic之类的函数来进行一般的打印、抛出错误处理。Go目前标准包只是包含了简单的功能，如果我们想把我们的应用日志保存到文件，然后又能够结合日志实现很多复杂的功能（编写过Java或者C++的读者应该都使用过log4j和log4cpp之类的日志工具），可以使用第三方开发的日志系统:logrus和seelog，它们实现了很强大的日志功能，可以结合自己项目选择。接下来我们介绍如何通过该日志系统来实现我们应用的日志功能。

### logrus介绍

logrus是用Go语言实现的一个日志系统，与标准库log完全兼容并且核心API很稳定,是Go语言目前最活跃的日志库

```
go get -u github.com/sirupsen/logrus
```

```
package main

import (
	log "github.com/Sirupsen/logrus"
)

func main() {
	log.WithFields(log.Fields{
		"animal": "walrus",
	}).Info("A walrus appears")
}
```

Supervisord是用Python实现的一款非常实用的进程管理工具。supervisord会帮你把管理的应用程序转成daemon程序，而且可以方便的通过命令开启、关闭、重启等操作，而且它管理的进程一旦崩溃会自动重启，这样就可以保证程序执行中断后的情况下有自我修复的功能。

## 小结

本章讨论了如何部署和维护我们开发的Web应用相关的一些话题。这些内容非常重要，要创建一个能够基于最小维护平滑运行的应用，必须考虑这些问题。

- 创建一个强健的日志系统，可以在出现问题时记录错误并且通知系统管理员
- 处理运行时可能出现的错误，包括记录日志，并如何友好的显示给用户系统出现了问题
- 处理404错误，告诉用户请求的页面找不到
- 将应用部署到一个生产环境中(包括如何部署更新)
- 如何让部署的应用程序具有高可用
- 备份和恢复文件以及数据库

