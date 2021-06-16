package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/ip2location/ip2location-go/v9"
	"github.com/umputun/go-flags"
	"github.com/umputun/reproxy/lib"
)

var version = "unknown"

var opts struct {
	Listen         string `short:"l" long:"listen" env:"LISTEN" description:"listen on host:port" default:"0.0.0.0:8080"`
	ReproxyAddress string `short:"r" long:"reproxy" env:"REPROXY" description:"reproxy plugins endpoint" default:"http://127.0.0.1:8081"`
	DatabasePath   string `short:"d" long:"database" env:"DATABASE" description:"database file path"`
	Fields         string `short:"f" long:"fields" env:"FIELDS" description:"fields string, comma-separated. See allowed values at https://github.com/negasus/reproxy-ip2location-plugin" default:"CF,CC,REG,CITY"`
	HeaderPrefix   string `short:"p" long:"prefix" env:"PREFIX" description:"http header prefix" default:"X-Geo-"`
	IPSource       string `short:"s" long:"ipsource" env:"IPSOURCE" description:"if defined, the remote address will be taken from that http header. For example: X-Real-IP" default:""`
	InMemory       bool   `short:"m" long:"inmemory" env:"INMEMORY" description:"if true, the database file will be loaded to the application memory"`
}

func main() {
	fmt.Printf("reproxy-ip2location-plugin %s\n", version)

	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.SubcommandsOptional = true
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			log.Printf("[ERROR] cli error: %v", err)
		}
		os.Exit(2)
	}

	log.Printf("options: %#v", opts)

	err := run()
	if err != nil {
		log.Printf("run plugin failed, %v", err)
		os.Exit(1)
	}

	log.Printf("done")
}

type dbReader struct {
	buf *bytes.Reader
}

func (b *dbReader) ReadAt(p []byte, off int64) (n int, err error) {
	return b.buf.ReadAt(p, off)
}

func (b *dbReader) Read(p []byte) (n int, err error) {
	return b.buf.Read(p)
}

func (b *dbReader) Close() error {
	return nil
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := &Handler{
		headerPrefix: opts.HeaderPrefix,
		ipSource:     opts.IPSource,
	}
	err := h.parseFields(opts.Fields)
	if err != nil {
		return err
	}

	var reader ip2location.DBReader

	if opts.InMemory {
		fileData, err := os.ReadFile(opts.DatabasePath)
		if err != nil {
			return err
		}
		reader = &dbReader{buf: bytes.NewReader(fileData)}
	} else {
		reader, err = os.Open(opts.DatabasePath)
		if err != nil {
			return err
		}
	}

	h.db, err = ip2location.OpenDBWithReader(reader)
	if err != nil {
		return err
	}
	defer h.db.Close()

	go func() {
		if x := recover(); x != nil {
			panic(x)
		}

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		cancel()
	}()

	plugin := lib.Plugin{
		Name:    "ip2location",
		Address: opts.Listen,
		Methods: []string{"HeaderThing"},
	}

	return plugin.Do(ctx, opts.ReproxyAddress, h)
}

type Handler struct {
	db           *ip2location.DB
	headerPrefix string
	ipSource     string
	fields       struct {
		countryFull        bool
		countryCode        bool
		region             bool
		city               bool
		isp                bool
		latitude           bool
		longitude          bool
		domain             bool
		zipcode            bool
		timezone           bool
		netspeed           bool
		iddcode            bool
		areacode           bool
		weatherstationcode bool
		weatherstationname bool
		mcc                bool
		mnc                bool
		mobilebrand        bool
		elevation          bool
		usagetype          bool
	}
}

func (h *Handler) parseFields(fields string) error {
	for _, f := range strings.Split(fields, ",") {
		switch f {
		case "CF":
			h.fields.countryFull = true
		case "CC":
			h.fields.countryCode = true
		case "REG":
			h.fields.region = true
		case "CITY":
			h.fields.city = true
		case "ISP":
			h.fields.isp = true
		case "LAT":
			h.fields.latitude = true
		case "LON":
			h.fields.longitude = true
		case "DOM":
			h.fields.domain = true
		case "ZIP":
			h.fields.zipcode = true
		case "TZ":
			h.fields.timezone = true
		case "NS":
			h.fields.netspeed = true
		case "IDD":
			h.fields.iddcode = true
		case "AREA":
			h.fields.areacode = true
		case "WEC":
			h.fields.weatherstationcode = true
		case "WEN":
			h.fields.weatherstationname = true
		case "MCC":
			h.fields.mcc = true
		case "MNC":
			h.fields.mnc = true
		case "MB":
			h.fields.mobilebrand = true
		case "EL":
			h.fields.elevation = true
		case "UT":
			h.fields.usagetype = true
		default:
			return fmt.Errorf("field %s not support", f)
		}
	}
	return nil
}

// by default, Handler get remote addr from request.RemoteAddr
// if Handler.ipSource is defined, handler get remote addr from request http header with that name
func (h *Handler) getIP(req lib.Request) (string, error) {
	var src string
	var err error

	if h.ipSource != "" {
		src = req.Header.Get(h.ipSource)
	}

	if src == "" {
		src, _, err = net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			return "", err
		}
	}

	ip := net.ParseIP(src)

	if ip == nil {
		return "", fmt.Errorf("wrong ip")
	}

	return ip.String(), nil
}

func (h *Handler) HeaderThing(req lib.Request, res *lib.Response) (err error) {
	ip, err := h.getIP(req)
	if err != nil {
		return fmt.Errorf("error get ip, %w", err)
	}

	rec, err := h.db.Get_all(ip)
	if err != nil {
		return fmt.Errorf("error get ip2location data for address %s, %w", ip, err)
	}

	res.HeadersIn = http.Header{}

	if h.fields.countryFull {
		res.HeadersIn.Add(h.headerPrefix+"Country", rec.Country_long)
	}
	if h.fields.countryCode {
		res.HeadersIn.Add(h.headerPrefix+"Country-Code", rec.Country_short)
	}
	if h.fields.region {
		res.HeadersIn.Add(h.headerPrefix+"Region", rec.Region)
	}
	if h.fields.city {
		res.HeadersIn.Add(h.headerPrefix+"City", rec.City)
	}
	if h.fields.isp {
		res.HeadersIn.Add(h.headerPrefix+"Isp", rec.Isp)
	}
	if h.fields.latitude {
		res.HeadersIn.Add(h.headerPrefix+"Latitude", strconv.FormatFloat(float64(rec.Latitude), 'f', 8, 64))
	}
	if h.fields.longitude {
		res.HeadersIn.Add(h.headerPrefix+"Longitude", strconv.FormatFloat(float64(rec.Longitude), 'f', 8, 64))
	}
	if h.fields.domain {
		res.HeadersIn.Add(h.headerPrefix+"Domain", rec.Domain)
	}
	if h.fields.zipcode {
		res.HeadersIn.Add(h.headerPrefix+"Zipcode", rec.Zipcode)
	}
	if h.fields.timezone {
		res.HeadersIn.Add(h.headerPrefix+"Timezone", rec.Timezone)
	}
	if h.fields.netspeed {
		res.HeadersIn.Add(h.headerPrefix+"Netspeed", rec.Netspeed)
	}
	if h.fields.iddcode {
		res.HeadersIn.Add(h.headerPrefix+"Iddcode", rec.Iddcode)
	}
	if h.fields.areacode {
		res.HeadersIn.Add(h.headerPrefix+"Areacode", rec.Areacode)
	}
	if h.fields.weatherstationcode {
		res.HeadersIn.Add(h.headerPrefix+"Weatherstationcode", rec.Weatherstationcode)
	}
	if h.fields.weatherstationname {
		res.HeadersIn.Add(h.headerPrefix+"Weatherstationname", rec.Weatherstationname)
	}
	if h.fields.mcc {
		res.HeadersIn.Add(h.headerPrefix+"Mcc", rec.Mcc)
	}
	if h.fields.mnc {
		res.HeadersIn.Add(h.headerPrefix+"Mnc", rec.Mnc)
	}
	if h.fields.mobilebrand {
		res.HeadersIn.Add(h.headerPrefix+"Mobilebrand", rec.Mobilebrand)
	}
	if h.fields.elevation {
		res.HeadersIn.Add(h.headerPrefix+"Elevation", strconv.FormatFloat(float64(rec.Elevation), 'f', 4, 64))
	}
	if h.fields.usagetype {
		res.HeadersIn.Add(h.headerPrefix+"Usagetype", rec.Usagetype)
	}

	return
}
