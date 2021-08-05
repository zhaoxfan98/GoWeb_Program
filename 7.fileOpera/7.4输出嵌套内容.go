package main

import (
	"os"
	"text/template"
)

type Friend struct {
	Fname string
}

type Person struct {
	UserName string
	Emails   []string
	Friends  []*Friend
}

func main5() {
	f1 := Friend{Fname: "zhaoxfan1"}
	f2 := Friend{Fname: "zhaoxfan2"}
	t := template.New("fieldname example")
	t, _ = t.Parse(`hello {{.UserName}}!
						{{range .Emails}}
							an email {{.}}
						{{end}}
						{{with .Friends}}
						{{range .}}
							my friend name is {{.Fname}}
						{{end}}
						{{end}}`)
	p := Person{UserName: "Zhaoxfan",
		Emails:  []string{"zhaoxfan@gmail.com", "18810617098@163.com"},
		Friends: []*Friend{&f1, &f2}}
	t.Execute(os.Stdout, p)
}
