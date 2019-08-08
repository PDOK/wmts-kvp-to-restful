package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	operations "github.com/PDOK/wmts-kvp-to-restful/operations"
)

// https://ndersson.me/post/capturing_status_code_in_net_http/
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewLoggingResponseWriter wrapper for the http.ResponseWriter
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-exists/10510718
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		log.Fatal(err)
		return false
	}
	log.Println(err)
	return true
}

func main() {

	host := flag.String("host", "http://localhost", "Hostname to proxy with protocol, http/https and port")
	template := flag.String("t", "", "Optional GetCapabilities template file, if not set request will be proxied.")
	logrequest := flag.Bool("l", false, "Enable request logging, default: false")
	flag.Parse()

	if len(*host) == 0 {
		log.Fatal("No target host is configured")
		return
	}

	if len(*template) > 0 && !exists(*template) {
		return
	}

	config := &operations.Config{Host: *host, Template: *template, Logging: *logrequest}

	origin, _ := url.Parse(*host)

	director := func(req *http.Request) {
		req.URL.Host = origin.Host
		req.URL.Scheme = origin.Scheme
		req.Host = origin.Host
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
	}

	proxy := &httputil.ReverseProxy{Director: director}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(`{"health": "OK"}`))
		return
	})

	log.Println("wmts-kvp-to-restful started")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		var logrequesturi string
		var start time.Time
		var elapsed time.Duration

		if config.Logging {
			start = time.Now()
			logrequesturi = r.URL.RequestURI()
		}

		lrw := newLoggingResponseWriter(w)
		mustproxy := operations.ProcessRequest(config, lrw, r)
		if mustproxy {
			proxy.ServeHTTP(lrw, r)

		}

		if config.Logging {
			elapsed = time.Since(start)
			if mustproxy {
				log.Printf("%d %s %s %s", lrw.statusCode, elapsed.Round(time.Millisecond), logrequesturi, r.URL.RequestURI())
			} else {
				log.Printf("%d %s %s", lrw.statusCode, elapsed.Round(time.Millisecond), r.URL.RequestURI())
			}
		}
		return
	})

	log.Fatal(http.ListenAndServe(":9001", nil))
}
