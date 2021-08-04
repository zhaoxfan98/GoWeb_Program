package main

import (
	"html/template"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	t := template.New("some template")       //创建一个模板
	t, _ = t.ParseFiles("tmpl/welcome.html") //解析模板文件
	user := GetUser()                        //获取当前用户信息
	t.Execute(w, user)                       //执行模板的merge操作
}

type Person struct {
	UserName string
}

func main4() {
	t := template.New("filename example")
	t, _ = t.Parse("hello {{.UserName}}!")
	p := Person{UserName: "Zhaoxfan"}
	t.Execute(os.Stdout, p)
}
