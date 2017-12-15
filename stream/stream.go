package stream

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"net/http"
	"io/ioutil"
	"github.com/ziutek/gst"
	"os"
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

//type Streams []Stream
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




	//stream.SName = resp.Header.Get("icy-name")
	//stream.SBitr = resp.Header.Get("icy-br")
	//stream.SGnr = resp.Header.Get("icy-genre")
	//stream.SMeta = extractMetadata(resp.Body, amount)

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



