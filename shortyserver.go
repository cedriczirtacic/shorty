package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"time"
)

type ShortyServer struct {
	port        int          // shorty server port
	address     string       // shorty server address
	starttime   time.Time    // when this server started?
	urllifespan int64        // the URL lifespan
	idlen       int          // length of ID
	sh          *ShortyHTTPd // pointer to ShortyHTTPd in use
	log         *ShortyLog   // pointer to logger
}

/* process_conn processes the net.Conn of a new client and creates
 * a new ShortyURL structure with users' supplied URL
 */
func (s *ShortyServer) process_conn(l net.Listener) {
	/* channel for short urls */
	urls := make(chan ShortyURL)

	for {
		c, err := l.Accept()
		if err != nil { /* got an error? */
			s.log.Printf("error: %s", err.Error())
			err = nil
			continue
		}

		/* data buffer */
		var buffer bytes.Buffer
		/* remote client info */
		var remoteaddr net.Addr

		remoteaddr = c.RemoteAddr()
		s.log.Printf("processing connection from %s\n", remoteaddr.String())
		i, err := io.Copy(&buffer, c)

		if i <= 1 {
			s.log.PrintErr("got null or no data from client\n")
			continue
		}
		if err != nil { /* got an error? */
			s.log.PrintErr(err.Error())
			err = nil
			continue
		}

		data := buffer.Bytes()
		/* trim newline if found */
		for {
			char := data[len(data)-1]
			if char == '\r' || char == '\n' {
				data = data[:len(data)-1]
				continue
			}
			break
		}

		/* process the url to generate a shorter one */
		go s.process_url(string(data), urls)

		url := <-urls
		/* we add the newly created short url */
		if url.err == nil {
			short_urls = append(short_urls, url)
		} else {
			s.log.PrintErr(url.err.Error())
		}
		s.log.Printf("generated URL: %s", url.URL(s.sh))

		/* pass short url to user */
		c.Write([]byte(url.URL(s.sh)))
		c.Close()

		// debug
		s.log.Printf("%v", short_urls)
	}
}

/* process_url processes the URL given by user and validates the data.
 */
func (s *ShortyServer) process_url(url string, c chan ShortyURL) (err error) {
	short := ShortyURL{
		timestamp: time.Now(),
	}
	// pre-regex validation
	if url[0:4] != "http" && url[0:5] != "https" {
		short.err = errors.New(fmt.Sprintf("invalid URL: %s", url))
		c <- short
		return
	}

	// regex validation
	re := regexp.MustCompile("^(?:http(s)?:\\/\\/)?[\\w.-]+(?:\\.[\\w\\.-]+)+[\\w\\-\\._~:/?#[\\]@!\\$&'\\(\\)\\*\\+,;=.]+$")
	if !re.MatchString(url) {
		short.err = errors.New(fmt.Sprintf("invalid URL: %s", url))
		c <- short
		return
	}

	/* that's the original url */
	short.src = url
	short.shorten(url, s.idlen)
	c <- short
	return
}
