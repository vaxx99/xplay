package fstx

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/xpath"
)

//A lll
type A struct {
	A, B, C, D, E string
}

//B lll
type B struct {
	A [20]A
}

func (s *B) proc(i int, fn, sf string) *B {
	switch fn {
	case "A":
		s.A[i].A = sf
	case "B":
		s.A[i].B = sf
	case "C":
		s.A[i].C = sf
	case "D":
		s.A[i].D = sf
	case "E":
		s.A[i].E = sf
	}
	return s
}

func (ss *B) Hproc() string {
	const tpl = `<html>
<head>
<title>FS.TO!</title>
<script src="./src/refresh.js"></script>
<link href="./css/style.css" rel="stylesheet">
</head>
<body>
<h1>Новое на портале FS.TO!</h1>
<table>`
	clr := map[string]string{
		"Сериалы": "RGB(255,250,205)",
		"Фильмы":  "RGB(135,206,250)",
		"0":       "RGB(248,248,255)",
	}

	s := tpl
	tr := ""

	for _, x := range ss.A {
		if len(clr[x.B]) > 0 {
			tr = "<tr style=\"background-color:" + clr[x.B] + "\">"
		} else {
			tr = "<tr style=\"background-color:" + clr["0"] + "\">"
		}

		s += tr + "<td>" + x.A + "</td><td>" + x.C + "</td><td>" + x.B + "</td><td>" + x.E + "</td><td>" + x.D + "</td></tr>\n"
	}
	s += "</table></body></html>\n"
	return s
}

func (a *B) Mproc() *B {
	resp, e := http.Get("http://fs.to")
	Ex(e)
	page, e := ioutil.ReadAll(resp.Body)
	Ex(e)

	// parse the web page
	doc, e := gokogiri.ParseHtml(page)
	Ex(e)

	//Xpath source
	s1 := "//*[@class=\"b-main__new-item-add-info-time\"]/text()"
	s2 := "//*[@class=\"b-main__new-item-subsection\"]/text()"
	s3 := "//*[@class=\"b-main__new-item-subsection-title\"]/span/text()"
	s4 := "//*[@class=\"b-main__new-item-attributes\"]"

	xp1 := xpath.Compile(s1)
	xp2 := xpath.Compile(s2)
	xp3 := xpath.Compile(s3)
	xp4 := xpath.Compile(s4)

	r0, _ := doc.Root().Search(xp1)
	r1, _ := doc.Root().Search(xp2)
	r2, _ := doc.Root().Search(xp3)
	r3, _ := doc.Root().Search(xp4)

	//Theme, Country
	for i := range r3 {
		if r3[i].CountChildren() == 5 {
			t := strings.Replace(r3[i].Content(), "\n", "", 1)
			s := strings.Replace(t, " ", "", -1)
			j := strings.Index(s, "\n")
			s0 := s[:j]
			s1 := s[j:]
			s1 = strings.Replace(s1, "\n", "", -1)
			a.proc(i, "D", s1)
			a.proc(i, "E", s0)

		} else {
			s := r3[i].Content()
			j := strings.Index(s, "\n")
			s = strings.Replace(s, "\n", "", -1)
			s = strings.Replace(s, " ", "", -1)
			if j == 0 {
				s = s[j:]
				a.proc(i, "D", s)
				a.proc(i, "E", "")
			} else {
				s = s[:j]
				a.proc(i, "D", "")
				a.proc(i, "E", s)
			}
		}
	}

	//Time
	for i, xx := range r0 {
		a.proc(i, "A", strings.Replace(xx.String(), "сегодня", "", -1))
	}

	//Type
	for i, xx := range r1 {
		a.proc(i, "B", xx.String())
	}

	//Name
	j := 0
	for i, xx := range r2 {
		if i%2 != 0 {
			a.proc(j, "C", xx.String())
			j++
		}
	}

	defer doc.Free()
	return a
}

// Ex - error handler
func Ex(e error) {
	if e != nil {
		fmt.Println("Error:", e)
		os.Exit(3)
	}
}
