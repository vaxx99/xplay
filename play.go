package main

import (
	"github.com/gotk3/gotk3/gtk"
	"log"
	"github.com/vaxx99/play/stream"
	"os"
	"time"
	"io"
	"bufio"
	"net/url"
	"net/http"
	"net"
	"fmt"
	"strings"
	"github.com/gotk3/gotk3/pango"
	"sync"
)

type Labels struct {
	Count, Station, Bitrate, Genre, Timer, Title *gtk.Label
}

func main() {
	gtk.Init(nil)
	proxy := ""
	os.Setenv("http_proxy", proxy)

	check := func(e error) {
		if e != nil {
			log.Fatalln(e)
		}
	}

	win, e := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	check(e)
	win.SetSizeRequest(400, 20)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	provider, _ := gtk.CssProviderNew()
	provider.LoadFromData(`@define-color base #333;@define-color fore #ccc;@define-color diff #555;
*{background-color: @base;border:0;padding: 5px;}
.container{color: @fore;border:0;}
.plus{color: @fore;border:0;font-size:20px;}
.station{color:@fore;border:0;font-size:10px;}
.bitrate{color:#ccc;border:0;font-size:10px;}
.timer{color:#ccc;border:0;font-size:10px;}
.title{color:#cff;border:0;font-size:10px;font-style: italic;}
.genre{color:#fcc;border:0;font-size:10px;}`)
	//e=provider.LoadFromPath("play.css")
	//c,e:=provider.ToString()
	//fmt.Println(e,c)
	//check(e)
	//header,_ := gtk.HeaderBarNew()
	//header.SetTitle("Streamer")
	//header.SetSubtitle("header")
	//header.SetShowCloseButton(true)
	//win.SetTitlebar(header)
	win.SetTitle("x:Player")
	win.SetIconFromFile("fsto.png")
	win.SetDefaultSize(400, 20)
	win.SetResizable(false)
	win.SetKeepAbove(true)

	cont, _ := gtk.GridNew()
	cont.SetRowHomogeneous(true)
	cont.SetSizeRequest(400, 20)
	plus, _ := gtk.LabelNew("")
	station, _ := gtk.LabelNew("")
	bitrate, _ := gtk.LabelNew("")
	genre, _ := gtk.LabelNew("")
	title, _ := gtk.LabelNew("")
	timer, _ := gtk.LabelNew("00:00:00")

	var labs = &Labels{plus, station, bitrate, genre, timer, title}

	ebox, _ := gtk.EventBoxNew()
	ubox, _ := gtk.EventBoxNew()
	cbox, _ := gtk.EventBoxNew()

	Style(provider, &cont.Widget, "container")
	Style(provider, &station.Widget, "station")
	Style(provider, &bitrate.Widget, "bitrate")
	Style(provider, &genre.Widget, "genre")
	Style(provider, &title.Widget, "title")
	Style(provider, &timer.Widget, "timer")
	Style(provider, &plus.Widget, "plus")

	plus.SetSizeRequest(20, 10)
	station.SetSizeRequest(300, 10)
	station.SetXAlign(0)
	bitrate.SetSizeRequest(40, 10)
	bitrate.SetXAlign(0)
	timer.SetSizeRequest(30, 10)
	timer.SetXAlign(0)
	title.SetSizeRequest(300, 10)
	title.SetXAlign(0)


	station.SetMaxWidthChars(10)
	station.SetEllipsize(pango.ELLIPSIZE_END)
	station.SetJustify(gtk.JUSTIFY_LEFT)

	title.SetMaxWidthChars(10)
	title.SetEllipsize(pango.ELLIPSIZE_END)
	title.SetJustify(gtk.JUSTIFY_LEFT)
	title.SetVisible(false)

	bitrate.SetMaxWidthChars(10)
	bitrate.SetEllipsize(pango.ELLIPSIZE_END)
	bitrate.SetJustify(gtk.JUSTIFY_LEFT)



	ebox.Add(timer)
	ubox.Add(station)
	cbox.Add(plus)

	cont.Attach(cbox, 0, 0, 1, 1)
	cont.Attach(ubox, 1, 0, 1, 1)
	cont.Attach(title, 1, 0, 1, 1)
	cont.Attach(bitrate, 2, 0, 1, 1)
	cont.Attach(ebox, 3, 0, 1, 1)


	var ch = make(chan struct{})
	var run = 0
	var play = stream.NewPlayer()

	//var cplay = stream.Stream{}
	//tmp = append(tmp,stream.Stream{Next:0,SName: "Radio Swiss Jazz", SUrl: "http://stream.srg-ssr.ch/m/rsj/aacp_96", SProxy: proxy})
	tmp := stream.Plist()
	cplay := tmp[0]

	labs.Count.SetMarkup("<span foreground='#f00'>•</span>")
	labs.Station.SetText(cplay.SName)

	ubox.Connect("event", func() {
		if labs.Station.GetVisible()==false{
			labs.Station.SetVisible(true)
			labs.Title.SetVisible(false)
		}else{
			labs.Station.SetVisible(false)
			labs.Title.SetVisible(true)
		}
	})

	ebox.Connect("button-press-event", func() {
		hb := &http.Response{}
		ha := 0
		titl := &Titles{Lab: labs}
		timr:=&Timer{time.Now(),labs}
		var wg sync.WaitGroup
		if run == 0 {
			run = 1
			uri,prx:=cplay.SUrl,cplay.SProxy
			hb,ha=TryOne(uri,prx)
			if ha == 0{
				hb,ha=TryTwo(uri,prx)
			}
			labs.Station.SetText(cplay.SName)
			if ha>0{
				data:=&Data{hb.Body,ha}
				titl.Now = &timr.Now
				cplay.SBitr = hb.Header.Get("icy-br")

				titl.Tch = data.Update(&wg,ch)
				labs.Station.SetVisible(false)
				labs.Title.SetVisible(true)
				wg.Add(1)
				go titl.Update(&wg,ch)
			}else{
				labs.Title.SetVisible(false)
			}
				labs.Count.SetMarkup("<span foreground='#0f0'>•</span>")
				labs.Bitrate.SetMarkup("<span foreground='#ff0'>" + cplay.SBitr + "</span><span foreground='#cfc'> K</span>")
				wg.Add(1)
				go timr.Update(&wg,ch)
				play.Player(cplay.SUrl)
		} else {
			run = 0
			labs.Clear()
			labs.Count.SetMarkup("<span foreground='#f00'>•</span>")
			play.Player("")
			ch <- struct{}{}
			ch <- struct{}{}
			labs.Title.SetVisible(false)
			labs.Station.SetVisible(true)
			wg.Wait()
		}
	})

	cbox.Connect("button-press-event", func() {
		cplay = tmp[cplay.Next]
		labs.Clear()
		labs.Station.SetText(cplay.SName)
		labs.Count.SetMarkup("<span foreground='#f00'>•</span>")
	})

	ubox.Connect("button-press-event", func() {
		tmp = stream.Plist()
		labs.Clear()
		labs.Station.SetText(cplay.SName)
		labs.Bitrate.SetText("•")
	})
	win.Add(cont)
	win.ShowAll()
	gtk.Main()
}

func Style(p *gtk.CssProvider, w *gtk.Widget, cn string) error {
	ctx, err := w.GetStyleContext()
	if err != nil {
		return err
	}
	ctx.AddProvider(p, 600)
	ctx.AddClass(cn)
	return nil
}

type Timer struct {
	Now time.Time
	Lab *Labels
}

func (t *Timer) Update(wg *sync.WaitGroup,ch chan struct{}) {
	for {
		select {
		default:
			time.Sleep(1 * time.Second)
			dur := time.Since(t.Now)
			tm, _ := time.Parse("150405", "000000")
			tt := tm.Add(dur)
			t.Lab.Timer.SetMarkup("<span foreground='#0f0'>" + tt.Format("15:04:05") + "</span>")
		case <-ch:
			t.Lab.Timer.SetText("00:00:00")
			//fmt.Println("1 wg done")
			wg.Done()
			return
		}
	}
}

type Titles struct {
	Tch <-chan string
	Lab *Labels
	Now *time.Time
}

func (t *Titles) Update(wg *sync.WaitGroup,ch chan struct{}) {
	select {
	default:
		for meta := range t.Tch {
				*t.Now = time.Now()
				t.Lab.Title.SetText(meta)
		}
		//fmt.Println("2 wg done")
	case <-ch:
		wg.Done()
		return
	}
}

func (L *Labels) Clear() {
	L.Bitrate.SetText("")
	L.Title.SetText("")
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

type Data struct{
	rdr io.Reader
	skip int
}

func (d *Data) Update(wg *sync.WaitGroup,ch chan struct{})<-chan string {
	dh := make(chan string)
	wg.Add(1)
	go func() {
		bufrdr := bufio.NewReaderSize(d.rdr, d.skip)
					for {
						skipbytes := make([]byte, d.skip)
						_, err := io.ReadFull(bufrdr, skipbytes)
						if err != nil {
							log.Printf("Failed: %v\n", err)
							close(dh)
							return
						}
						c, err := bufrdr.ReadByte()
						if err != nil {
							log.Panic(err)
						}
						if c > 0 {
							meta, err := ParseIcy(bufrdr, c)
							if err != nil {
								log.Panic(err)
							}
							dh <- meta
						}
					select {
					default:
					case <-ch:
						//fmt.Println("Return", d.skip)
						wg.Done()
						//fmt.Println("3 wg done")
						close(dh)
						return
					}
				}
		}()
	return dh
}

func TryOne(uri,proxy string) (*http.Response,int) {
	proxyUrl, err := url.Parse(proxy)
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl)}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return &http.Response{},0
	}
	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		return nil,0
	}
	amount := 0
	if _, err = fmt.Sscan(resp.Header.Get("icy-metaint"), &amount); err != nil {
		return nil,0
	}
	return resp, amount
}

func TryTwo(uri,proxy string) (*http.Response,int){
	proxyUrl, err := url.Parse(proxy) //!
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl), //!
		Dial: func(network, a string) (net.Conn, error) {
			realConn, err := net.Dial(network, a)
			if err != nil {
				return nil,err
			}
			return &IcyCW{Conn: realConn}, nil
		},
	}
	client := &http.Client{Transport: tr}
	http.DefaultClient = client
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, 0
	}
	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	amount := 0
	if _, err = fmt.Sscan(resp.Header.Get("icy-metaint"), &amount); err != nil {
		return nil,0
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

