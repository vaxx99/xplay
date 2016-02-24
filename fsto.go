package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const tpl = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>FS.TO!</title>
<style>
body {  width: 100%;
	height: 100%;
	margin-top: 0;
	margin-left: 0;
	margin-right: 0;
	margin-bottom: 0;
	padding: 0;
	background: #4e4747; }
table {
	width: 100%;
	border-collapse: collapse;
	font-family: 'Arial';
	font-size: 8pt;
	}
th {
	padding: 2px;
	border: 1px solid black;
	font-family: 'Arial';
	font-size: 8pt;
	}
td {
	padding: 2px;
	border: 1px solid black;
	font-family: 'Arial';
	font-size: 8pt;
	}
h1 {
	margin: 0;
	color: rgb(140, 140, 140);
	font-family: 'Arial';
	font-size: 12pt;
	text-align: center;
	}
</style>
<script>
<!--
function autoRefresh()
{
window.location = window.location.href;
}
setInterval('autoRefresh()', 600000);
//-->
</script>
</head>
<body>
<h1>Новое на портале FS.TO!</h1>
<table>`

func page() string {

	clr := map[string]string{
		"Сериалы": "RGB(255,250,205)",
		"Фильмы":  "RGB(135,206,250)",
		"0":       "RGB(248,248,255)",
	}

	pg := tpl
	tr := ""

	doc, err := goquery.NewDocument("https://fs.to")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".b-main__new-item ").Each(func(i int, s *goquery.Selection) {
		t0 := s.Find(".b-main__new-item-add-info-time").Text()
		t1 := s.Find(".b-main__new-item-title ").Text()
		t2 := s.Find(".b-main__new-item-subsection").Text()
		t3 := s.Find(".b-main__new-item-attributes").Text()
		t4 := s.Find(".b-main__new-item-description").Text()
		t4 = strings.Replace(t4, "\n", "", 1)
		t4 = strings.Replace(t4, "\"", "", -1)
		t := strings.Replace(t3, "\n", "", 1)
		ss := strings.Replace(t, " ", "", -1)
		j := strings.Index(ss, "\n")
		s0 := ss[:j]
		s1 := ss[j:]
		s1 = strings.Replace(s1, "\n", "", -1)
		//color
		if len(clr[t2]) > 0 {
			tr = "<tr style=\"background-color:" + clr[t2] + "\">"
		} else {
			tr = "<tr style=\"background-color:" + clr["0"] + "\">"
		}
		//table
		pg += tr + "<td>" + t0 + "</td><td title=\"" + t4 + "\">" + t1 + "</td><td>" + t2 + "</td><td>" + s0 + "</td><td>" + s1 + "</td></tr>\n"
	})
	pg += "</table></body></html>\n"
	return pg
}

func Fsto(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s", page())
}

func main() {
	http.HandleFunc("/", Fsto)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
