package main

import (
	"log"
	"os"
	"time"
	"bufio"
	"io"
	"net/http"
	"strings"
	"sync"
	"os/exec"
	"strconv"
	"github.com/vaxx99/play/stream"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"github.com/gotk3/gotk3/gdk"
)

type Labels struct {
	Count, Station, Bitrate, Genre, Timer, Title *gtk.Label
}

func main() {
	gtk.Init(nil)

	check := func(m string, e error) {
		if e != nil {
			log.Println(m, e)
		}
	}

	win, e := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	check("", e)
	win.SetSkipTaskbarHint(true)

	//Linux screen resolution
	out, _ := exec.Command("xdpyinfo").Output()
	dpi := string(out)
	dpi = dpi[strings.Index(dpi, "dimensions:")+len("dimensions:") : strings.Index(dpi, "pixels")]
	a := strings.Split(dpi, "x")
	W, e := strconv.Atoi(strings.TrimSpace(a[0]))
	H, e := strconv.Atoi(strings.TrimSpace(a[1]))
	check("Resolution:", e)
	if e != nil {
		W, H = 1024, 768
	}

	//win.SetSizeRequest(300, 12)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	win.Connect("key-press-event",func(win *gtk.Window,ev *gdk.Event){
		keyEvent := &gdk.EventKey{ev}
		if keyEvent.KeyVal() == gdk.KEY_Escape{
			gtk.MainQuit()
		}

	})

	provider, _ := gtk.CssProviderNew()
	provider.LoadFromData(`@define-color base #333;@define-color fore #ccc;@define-color diff #555;*{background-color: @base;border:0;}.container{color: @fore;border:1px solid @base;}.plus{color: @fore;border:0;font-size:20px;}.station{color:@fore;border:0;font-size:10px;}.bitrate{color:#ccc;border:0;font-size:10px;}.timer{color:#ccc;border:0;font-size:10px;padding-right:4px;}.title{color:#cff;border:0;font-size:10px;}.genre{color:#fcc;border:0;font-size:10px;}`)
	win.SetTitle("xPlay")
	win.SetIconFromFile("fsto.png")
	win.SetDefaultSize(300, 12)
	win.SetResizable(false)
	win.SetKeepAbove(true)

	cont, _ := gtk.GridNew()
	cont.SetRowHomogeneous(true)
	cont.SetColumnSpacing(4)
	cont.SetSizeRequest(300, 12)
	cont.SetVExpand(false)

	plus, _ := gtk.LabelNew("")
	station, _ := gtk.LabelNew("")
	bitrate, _ := gtk.LabelNew("")
	genre, _ := gtk.LabelNew("")
	title, _ := gtk.LabelNew("")
	timer, _ := gtk.LabelNew("00:00")

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

	plus.SetSizeRequest(20, 12)
	station.SetSizeRequest(230, 12)
	station.SetXAlign(0)
	bitrate.SetSizeRequest(20, 12)
	bitrate.SetXAlign(-1)
	timer.SetSizeRequest(20, 12)
	timer.SetXAlign(-1)
	title.SetSizeRequest(230, 12)
	title.SetXAlign(0)

	station.SetMaxWidthChars(10)
	station.SetEllipsize(pango.ELLIPSIZE_END)
	station.SetJustify(gtk.JUSTIFY_LEFT)

	title.SetLines(2)
	title.SetMaxWidthChars(10)
	title.SetLineWrap(true)
	//title.SetEllipsize(pango.ELLIPSIZE_END)
	title.SetJustify(gtk.JUSTIFY_CENTER)
	title.SetVisible(false)

	bitrate.SetMaxWidthChars(10)
	bitrate.SetEllipsize(pango.ELLIPSIZE_END)
	bitrate.SetJustify(gtk.JUSTIFY_RIGHT)
	timer.SetJustify(gtk.JUSTIFY_RIGHT)

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
			hb, ha = stream.TryOne(uri, prx)
			if ha == 0 {
				hb, ha = stream.TryTwo(uri, prx)
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
				labs.Bitrate.SetMarkup("<span foreground='#ff0'>" + cplay.SBitr + "</span><span foreground='#cfc'>k</span>")
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
				x = 0
			}
			if y < H/2 {
				y = 0
			}
			if x > W/2 {
				x = W - w
			}
			if y > H/2 {
				y = H
			}
			win.Move(x, y)
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
			tm, _ := time.Parse("0405", "0000")
			tt := tm.Add(dur)
			t.Lab.Timer.SetMarkup("<span foreground='#0ff'>" + tt.Format("04:05") + "</span>")
		case <-ch:
			t.Lab.Timer.SetText("00:00")
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

type Data struct {
	rdr  io.Reader
	skip int
}

func (t *Titles) Update(wg *sync.WaitGroup, ch chan struct{}) {
	select {
	default:
		for meta := range t.Tch {
			*t.Now = time.Now()
			xtx:=toUtf8([]byte(meta))
			if len(xtx)>0{
				xtx=strings.TrimLeft(xtx,"'")
				xtx=strings.TrimRight(xtx,"'")
				if strings.Contains(xtx,"&"){
					xtx = strings.Replace(xtx, "&", "&amp;", -1)
				}
			}
			t.Lab.Title.SetMarkup("<span color='#cff' font-style='italic'>"+xtx+"</span>")
		}
	case <-ch:
		wg.Done()
		return
	}
}

func toUtf8(iso8859_1_buf []byte) string {
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

func (L *Labels) Clear() {
	L.Bitrate.SetText("")
	L.Title.SetText("")
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
				meta, err := stream.ParseIcy(bufrdr, c)
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

