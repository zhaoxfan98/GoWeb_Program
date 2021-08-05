package main

import (
	"fmt"
	"net"
	"os"
)

func main1() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage:%s host:port ", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]
	//首先将用户输入作为参数传入  获取一个tcpAddr
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	CheckError(err)
	//将tcpAddr传入DialTCP后创建了一个TCP连接conn
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	CheckError(err)
	//通过conn来发送请求信息
	_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	CheckError(err)
	//后通过ioutil.ReadAll从conn中读取全部的文本，也就是服务端响应反馈的信息。
	// result, err := ioutil.ReadAll(conn)
	result := make([]byte, 256)
	_, err = conn.Read(result)
	CheckError(err)
	fmt.Println(string(result))
	os.Exit(0)
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
