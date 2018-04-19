package main

import (
	"fmt"
	"net/http"
	"os"
)

type ShortyHTTPd struct {
	port    int           // shorty httpd port
	address string        // shorty httpd address
	sslcert string        // SSL/TLS certificate
	sslkey  string        // SSL/TLS private key
	ss      *ShortyServer // pointer ShortyServer
	log     *ShortyLog    // pointer to logger
}

/* process_request get the HTTP request and redirects if short URL
 * exists. If id is unexsitent responds with 404 and if invalid
 * responds with a 403.
 */
func (sh *ShortyHTTPd) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	/* only accept GET */
	if req.Method != "GET" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	/* clean URI */
	id := req.RequestURI[1:]
	for i, c := range id {
		switch c {
		case '/':
		case '?':
			id = id[:i]
			goto cont
		}
	}
cont:

	/* load a index.html in case of / or /index.html or plain
	 * response in case of non-existent file
	 */
	if len(id) == 0 || id == "index.html" {
		_, err = os.Stat("index.html")
		if /* err == nil { */ !os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/html")
			http.ServeFile(w, req, "index.html")
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fmt.Sprintf("Shorty service running on %s port %d",
				sh.ss.address, sh.ss.port)))
		}
	} else {
		if len(id) != sh.ss.idlen {
			/* invalid length */
			w.WriteHeader(http.StatusForbidden)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Wrong id."))
		} else {
			found, short := idexists(id)
			if !found {
				/* id doesn't exist */
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte("Not Found."))
				return
			}

			/* we found the short URL */
			http.Redirect(w, req, short.src, http.StatusFound)
		}
	}
}
