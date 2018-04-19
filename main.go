package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

/***
 * SHORT URL
 ***/
type ShortyURL struct {
	id        string    // id of short URL
	src       string    // real source of the URL
	err       error     // error
	timestamp time.Time // when was created
}

/* shorten generates the short URL id
 */
func (su *ShortyURL) shorten(u string, l int) {
	id := make([]byte, l)
	for {
		rand.Seed(time.Now().UnixNano())

		length := l /* length of id */
		for length > 0 {
			r := rand.Intn(len(id_chars))
			id[length-1] = id_chars[r]
			length--
		}

		/* is already there? */
		su.id = string(id)
		if e, _ := idexists(string(id)); !e {
			break
		}
	}
}

/* idexists recursively checks the short URL array and
 * and returns true in case of existance and the
 * *ShortyURL data, or false and nil in case of inexistance.
 */
func idexists(id string) (bool, *ShortyURL) {
	for _, s := range short_urls {
		if id == s.id {
			return true, &s
		}
	}
	return false, nil
}

/* URL returns a formatted URL to be used
 */
func (su *ShortyURL) URL(h *ShortyHTTPd) string {
	if h.port != 443 {
		return fmt.Sprintf("https://%s:%d/%s", h.domain, h.port, su.id)
	}
	return fmt.Sprintf("https://%s/%s", h.domain, su.id)
}

/* expiration checks for age of the short url and returns
 * a true in case of expiration or false if still usable.
 */
func (su *ShortyURL) expiration(span int64) bool {
	/* span will be treated as minutes */
	now := time.Now().Unix()
	age := (now - su.timestamp.Unix()) / 60

	if age >= span { /* this short url expired */
		return true
	}
	return false
}

/***
 * LOGGING
 ***/
type ShortyLog struct {
	*log.Logger
}

/* NewLogger creates a log.Logger interface to work with.
 */
func NewLogger(o io.Writer) *ShortyLog {
	l := log.New(o, "[shorty] ", log.LstdFlags|log.Lmicroseconds)
	return &ShortyLog{l}
}

/* PrintErr prints a custom error
 */
func (log *ShortyLog) PrintErr(fmt string, s ...interface{}) {
	if len(s) > 0 {
		log.Printf("[ERROR] "+fmt, s...)
	} else {
		log.Print("[ERROR] " + fmt)
	}
}

/* global variables */
var (
	err          error
	shorty       ShortyServer
	shorty_httpd ShortyHTTPd
	shortylog    *ShortyLog

	id_chars = []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	}

	short_urls []ShortyURL
)

/* append_address creates a string with format ADDRESS:PORT
 */
func append_address(a string, p int) string {
	i := strconv.Itoa(p)
	return fmt.Sprintf("%s:%s", a, i)
}

func init() {
	/* create main server structs */
	shortylog = NewLogger(os.Stderr)
	shorty = ShortyServer{log: shortylog}
	shorty_httpd = ShortyHTTPd{log: shortylog}

	shorty.sh = &shorty_httpd
	shorty_httpd.ss = &shorty

	/* parse flags */
	flag.IntVar(&shorty.port, "p", 6666, "Shorty server port.")
	flag.StringVar(&shorty.address, "h", "localhost", "Shorty server address.")
	flag.Int64Var(&shorty.urllifespan, "l", 10, "Short url lifespan.")
	flag.IntVar(&shorty.idlen, "L", 6, "Length of url ID.")
	flag.IntVar(&shorty_httpd.port, "wp", 443, "Shorty httpd server port.")
	flag.StringVar(&shorty_httpd.address, "wh", "localhost", "Shorty httpd server address.")
	flag.StringVar(&shorty_httpd.domain, "wd", "localhost", "Shorty httpd server domain to use.")

	/* SSL/TLS flags */
	flag.StringVar(&shorty_httpd.sslcert, "wcert", "shorty.crt", "Shorty httpd ssl certificate path.")
	flag.StringVar(&shorty_httpd.sslkey, "wkey", "shorty.key", "Shorty httpd ssl private key path.")

	flag.Parse()

	if !flag.Parsed() {
		flag.PrintDefaults()
		return
	}
}

func main() {
	signals := make(chan os.Signal, 1)

	/* catch these signals */
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGILL)

	/* create tcp(4,6) listener */
	tcp_listener, err := net.Listen("tcp", append_address(shorty.address, shorty.port))
	if err != nil {
		shortylog.Fatalf(err.Error())
	}
	defer tcp_listener.Close()
	/* setup starting time */
	shorty.starttime = time.Now()

	/* setup the https server */
	go func() {
		//process := http.HandlerFunc(shorty_httpd.process_request)
		shortylog.Printf("setting up httpd server on: https://%s:%d", shorty_httpd.domain,
			shorty_httpd.port)
		err = http.ListenAndServeTLS(append_address(shorty_httpd.address, shorty_httpd.port),
			shorty_httpd.sslcert,
			shorty_httpd.sslkey,
			&shorty_httpd)
		if err != nil {
			shortylog.Panic(err.Error())
		}
	}()

	/* this dude is going to monitor the expiration */
	go func(urls *[]ShortyURL, ss *ShortyServer) {
		for {
			for i, u := range *urls {
				if u.expiration(ss.urllifespan) {
					/* short url is going to die */
					*urls = append((*urls)[:i], (*urls)[i+1:]...)
					break
				}
			}
			//time.Sleep(1 * time.Second)
		}
	}(&short_urls, &shorty)

	/* catch the signals and die gracefully */
	go func() {
		sig := <-signals
		shortylog.PrintErr("Catched signal: %s", sig)

		os.Exit(1)
	}()

	/* main shorty server loop */
	shortylog.Printf("setting up the shorty server on: %s:%d", shorty.address,
		shorty.port)
	shorty.process_conn(tcp_listener)
}
