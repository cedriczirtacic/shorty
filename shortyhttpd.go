package main

import (
	"net/http"
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
func (sh *ShortyHTTPd) process_request(w http.ResponseWriter, req *http.Request) {
	/* only accept GET */
	if req.Method != "GET" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	id := req.RequestURI[1:]
	for i, c := range id {
		switch c {
		case '/':
		case '?':
			id = id[:i]
			break
		}
	}
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
