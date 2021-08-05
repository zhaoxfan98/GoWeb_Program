//Go如何处理login页面的form数据

package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

func login(w http.ResponseWriter, r *http.Request) {
	//默认情况下，Handler里面是不会自动解析form的，必须显式的调用r.ParseForm()后，才能对这个表单数据进行操作
	fmt.Println("method:", r.Method) //获取请求的方法
	if r.Method == "GET" {
		timestamp := strconv.Itoa(time.Now().Nanosecond())
		hashWr := md5.New()
		hashWr.Write([]byte(timestamp))
		token := fmt.Sprintf("%x", hashWr.Sum(nil))

		t, _ := template.ParseFiles("login.gtpl")
		//token根据时间不断变化，这样保证了每次显示form表单的时候都是唯一的，用户递交的表单保持了唯一性。
		t.Execute(w, token)
	} else {
		r.ParseForm()
		//请求的是登录数据，那么执行登录的逻辑判断
		token := r.Form.Get("token")
		if token != "" {
			//验证token的合法性
		} else {
			//不存在token报错
		}
		fmt.Println("username length:", len(r.Form["username"][0]))
		fmt.Println("username:", template.HTMLEscapeString(r.Form.Get("username"))) //输出到服务器端
		fmt.Println("password:", template.HTMLEscapeString(r.Form.Get("password")))
		template.HTMLEscape(w, []byte(r.Form.Get("username"))) //输出到客户端
	}
}

func main2() {
	http.HandleFunc("/login", login)         //设置访问的路由
	err := http.ListenAndServe(":9090", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
