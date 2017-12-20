package stream

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"net/http"
	"io/ioutil"


	"github.com/ziutek/gst"

	"os"
	"bufio"
	"io"
	"log"
	"net/url"
	"fmt"
	"net"
)

type Stream struct {
	Crnt int
	Next int
	SName  string
	SProxy string
	SUrl   string
	SBitr  string
	SGnr   string
	STitle string
}

type Streams []Stream

func Plist() Streams {
	var tmp Streams
	proxy:=os.Getenv("http_proxy")
	tmp = append(tmp,Stream{SName:"Radio Swiss Jazz",SUrl: "http://stream.srg-ssr.ch/m/rsj/aacp_96",SProxy:proxy})
	j := len(tmp)
	tmp[0].Crnt = 0
	tmp[0].Next = j
	doc, _ := goquery.NewDocument("https://www.internet-radio.com")
	doc.Find(".text-danger").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")
		_, e := a.Attr("href")
		if e != false {
				c := a.Text()
			tmp = append(tmp,Stream{SName:c,SProxy:proxy})

		}
	})
	doc.Find("small.hidden-xs").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")
		b, e := a.Attr("onclick")
		if e != false {
			c := b[strings.Index(b, "http://") : strings.Index(b, ")")-1]
			resp, er := http.Get(c)
			if er != nil {
				panic(er)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			u := string(body)
			u = strings.Replace(u, "\n", "", -1)
			u = strings.Replace(u, "\r", "", -1)
			if strings.Index(u, "File1=") > 0 {
				u = u[strings.Index(u, "File1=")+len("File1="):]
			}
			if strings.Index(u, "Title") > 0 {
				u = u[:strings.Index(u, "Title")]
			}
			//lf := ""
			//if strings.Index(c, ".pls") > 0 {
			//	lf = ""
			//}
			//st[j+i].SID = j+i
			//st[j+i].SUrl = u
			tmp[j+i].Crnt = j+i
			tmp[j+i].Next = j+i+1
			tmp[j+i].SUrl = u
		}
	})
	tmp[len(tmp)-1].Next = 0
	return tmp
}

type Player struct {
	Play *gst.Element
}

func NewPlayer() *Player {
	return &Player{gst.ElementFactoryMake("playbin", "player")}
}

func (p *Player) Player(url string){
	state,_,_:=p.Play.GetState(100)
	if state == gst.STATE_NULL{
		p.Play.SetProperty("uri",url)
		p.Play.SetState(gst.STATE_PLAYING)
	}
	if state == gst.STATE_PLAYING{
		p.Play.SetProperty("uri","")
		p.Play.SetState(gst.STATE_NULL)
	}
}

func ParseIcy(rdr *bufio.Reader, c byte) (string, error) {
	numbytes := int(c) * 16
	bytes := make([]byte, numbytes)
	n, err := io.ReadFull(rdr, bytes)
	if err != nil {
		log.Panic(err)
	}
	if n != numbytes {
		return "", nil
	}
	return strings.Split(strings.Split(string(bytes), "=")[1], ";")[0], nil
}

func TryOne(uri, proxy string) (*http.Response, int) {
	client := &http.Client{}
	trans := &http.Transport{}
	proxyUrl, err := url.Parse(proxy)

	if proxy != "" {
		trans.Proxy = http.ProxyURL(proxyUrl)
		client = &http.Client{Transport: trans}
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return &http.Response{}, 0
	}
	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		return &http.Response{}, 0
	}
	amount := 0
	if _, err = fmt.Sscan(resp.Header.Get("icy-metaint"), &amount); err != nil {
		return &http.Response{}, 0
	}

	return resp, amount
}

func TryTwo(uri, proxy string) (*http.Response, int) {
	trans := &http.Transport{
		Dial: func(network, a string) (net.Conn, error) {
			realConn, err := net.Dial(network, a)
			if err != nil {
				return nil, err
			}
			return &IcyCW{Conn: realConn}, nil
		},
	}
	proxyUrl, err := url.Parse(proxy) //!
	if err == nil {
		trans.Proxy = http.ProxyURL(proxyUrl)
	}
	client := &http.Client{Transport: trans}
	http.DefaultClient = client
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return &http.Response{}, 0
	}
	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	amount := 0
	if _, err = fmt.Sscan(resp.Header.Get("icy-metaint"), &amount); err != nil {
		return &http.Response{}, 0
	}
	return resp, amount
}

// ICY - Metadata
type IcyCW struct {
	net.Conn
	haveReadAny bool
}

func (i *IcyCW) Read(b []byte) (int, error) {
	if i.haveReadAny {
		return i.Conn.Read(b)
	}
	i.haveReadAny = true
	n, err := i.Conn.Read(b[:3])
	if err != nil {
		return n, err
	}
	if string(b[:3]) == "ICY" {
		copy(b, []byte("HTTP/1.1"))
		return 8, nil
	}
	return n, nil
}



