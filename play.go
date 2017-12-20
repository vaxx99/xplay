package main

import (
	"github.com/gotk3/gotk3/gtk"
	"log"

	"github.com/vaxx99/play/stream"
	"os"

	"time"

	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gotk3/gotk3/pango"
	"sync"

	"os/exec"
	"strconv"
)

type Labels struct {
	Count, Station, Bitrate, Genre, Timer, Title *gtk.Label
}

func main() {
	gtk.Init(nil)

	check := func(m string,e error) {
		if e != nil {
			log.Println(m,e)
		}
	}

	win, e := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	check("",e)
	win.SetSkipTaskbarHint(true)

	//Linux screen resolution
	out,_:=exec.Command("xdpyinfo").Output()
	dpi:=string(out)
	dpi=dpi[strings.Index(dpi,"dimensions:")+len("dimensions:"):strings.Index(dpi,"pixels")]
	a:=strings.Split(dpi,"x")
	W,e:=strconv.Atoi(strings.TrimSpace(a[0]))
	H,e:=strconv.Atoi(strings.TrimSpace(a[1]))
	check("Resolution:",e)
	if e!=nil{
		W, H = 1024, 768
	}

	win.SetSizeRequest(400, 10)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	provider, _ := gtk.CssProviderNew()
	provider.LoadFromData(`@define-color base #333;@define-color fore #ccc;@define-color diff #555;
*{background-color: @base;border:0;}
.container{color: @fore;border:1px solid @base;}
.plus{color: @fore;border:0;font-size:20px;}
.station{color:@fore;border:0;font-size:10px;}
.bitrate{color:#ccc;border:0;font-size:10px;}
.timer{color:#ccc;border:0;font-size:10px;padding-right:5px;}
.title{color:#cff;border:0;font-size:10px;font-style: italic;}
.genre{color:#fcc;border:0;font-size:10px;}`)

	win.SetTitle("xPlay")
	win.SetIconFromFile("fsto.png")
	win.SetDefaultSize(400, 10)
	win.SetResizable(false)
	win.SetKeepAbove(true)


	cont, _ := gtk.GridNew()
	cont.SetRowHomogeneous(true)
	cont.SetSizeRequest(400, 10)

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
	bbox, _ := gtk.EventBoxNew()

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
	bbox.Add(bitrate)

	cont.Attach(cbox, 0, 0, 1, 1)
	cont.Attach(ubox, 1, 0, 1, 1)
	cont.Attach(title, 1, 0, 1, 1)
	cont.Attach(bbox, 2, 0, 1, 1)
	cont.Attach(ebox, 3, 0, 1, 1)

	var ch = make(chan struct{})
	var run = 0
	var play = stream.NewPlayer()
	var tmp = stream.Streams{}

	proxy := os.Getenv("http_proxy")

	if proxy == "" {
		tmp = append(tmp, stream.Stream{Crnt: 0, Next: 0, SName: "Radio Swiss Jazz", SUrl: "http://stream.srg-ssr.ch/m/rsj/aacp_96"})
	} else {
		os.Setenv("http_proxy", proxy)
		tmp = stream.Plist()
	}
	cplay := tmp[0]

	labs.Count.SetMarkup("<span foreground='#f00'>•</span>")
	labs.Station.SetText(cplay.SName)

	ubox.Connect("event", func() {
		if labs.Station.GetVisible() == false {
			labs.Station.SetVisible(true)
			labs.Title.SetVisible(false)
		} else {
			labs.Station.SetVisible(false)
			labs.Title.SetVisible(true)
		}
	})

	ebox.Connect("button-press-event", func() {
		hb := &http.Response{}
		ha := 0
		titl := &Titles{Lab: labs}
		timr := &Timer{time.Now(), labs}
		var wg sync.WaitGroup
		if run == 0 {
			run = 1
			uri, prx := cplay.SUrl, cplay.SProxy
			hb, ha = TryOne(uri, prx)
			if ha == 0 {
				hb, ha = TryTwo(uri, prx)
			}
			labs.Station.SetText(cplay.SName)
			if hb.StatusCode == 200 && ha > 0 {
				data := &Data{hb.Body, ha}
				titl.Now = &timr.Now
				cplay.SBitr = hb.Header.Get("icy-br")
				titl.Tch = data.Update(&wg, ch)
				labs.Station.SetVisible(false)
				labs.Title.SetVisible(true)
				wg.Add(1)
				go titl.Update(&wg, ch)
				labs.Count.SetMarkup("<span foreground='#0f0'>•</span>")
				labs.Bitrate.SetMarkup("<span foreground='#ff0'>" + cplay.SBitr + "</span><span foreground='#cfc'> k</span>")
				wg.Add(1)
				go timr.Update(&wg, ch)
				play.Player(cplay.SUrl)
			} else {
				run = 0
				labs.Title.SetVisible(false)
			}
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
		if proxy != "" {
			tmp = stream.Plist()
			labs.Clear()
			labs.Station.SetText(cplay.SName)
			labs.Count.SetMarkup("<span foreground='#ff0'>•</span>")
			cplay = tmp[0]
		}
	})

	//bbox.Connect("event", func() {
	//	bitrate.SetMarkup("<span background='#222'>move</span>")
	//})

	bbox.Connect("button-press-event", func() {
		x, y := win.GetPosition()
		w, h := win.GetSize()

		if win.GetDecorated() == true {
			win.SetDecorated(false)
			if x < W/2 {
				x=0
			}
			if y < H/2 {
				y=0
			}
			if x > W/2{
				x=W-w
			}
			if y > H/2 {
				y=H
			}
			win.Move(x,y)
		} else {
			win.Move(W/2-w/2, H/2-h/2)
			win.SetDecorated(true)
		}
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

func (t *Timer) Update(wg *sync.WaitGroup, ch chan struct{}) {
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

func (t *Titles) Update(wg *sync.WaitGroup, ch chan struct{}) {
	select {
	default:
		for meta := range t.Tch {
			*t.Now = time.Now()
			t.Lab.Title.SetText(meta)
		}
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

type Data struct {
	rdr  io.Reader
	skip int
}

func (d *Data) Update(wg *sync.WaitGroup, ch chan struct{}) <-chan string {
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
				wg.Done()
				close(dh)
				return
			}
		}
	}()
	return dh
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
