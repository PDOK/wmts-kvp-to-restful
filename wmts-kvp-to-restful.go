package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

func queryToPath(query map[string][]string) (path string, exception []byte) {
	var layer, tilematrixset, tilematrix, tilecol, tilerow, format string

	var regex = regexp.MustCompile(`^.*:(.*)$`)

	for key, value := range query {
		value := value[0]

		if strings.ToLower(key) == "layer" {
			layer = value
		}

		if strings.ToLower(key) == "tilematrixset" {
			tilematrixset = value
		}

		if strings.ToLower(key) == "tilematrix" {
			groups := regex.FindAllStringSubmatch(value, -1)

			if groups != nil {
				tilematrix = groups[0][1]
			} else {
				tilematrix = value
			}
		}

		if strings.ToLower(key) == "tilerow" {
			tilerow = value
		}

		if strings.ToLower(key) == "tilecol" {
			tilecol = value
		}

		if strings.ToLower(key) == "format" {
			if value == "image/png8" {
				format = ".png"
			} else if value == "image/jpeg" {
				format = ".jpeg"
			} else {
				format = ".png"
			}
		}
	}

	path = "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + format

	return path, nil
}

func buildNewPath(urlPath, newQueryPath string) string {
	return strings.TrimRight(urlPath, "/") + newQueryPath
}

func isTileQuery(query map[string][]string) bool {
	for _, key := range [6]string{
		"layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format",
	} {
		_, ok := query[key]
		if !ok {
			return false
		}
	}
	return true
}

func main() {

	host := flag.String("host", "http://localhost", "Hostname to proxy with protocol, http/https and port")

	flag.Parse()

	if len(*host) == 0 {
		log.Fatal("No target host is configured")
		return
	}

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
		query := r.URL.Query()

		if isTileQuery(query) {
			newpath, exception := queryToPath(query)
			if exception != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write([]byte(`{"status": "rewrite went wrong"}`))
				return
			}
			r.URL.Path = buildNewPath(r.URL.Path, newpath)
			r.URL.RawQuery = ""
		}

		proxy.ServeHTTP(w, r)
		return
	})

	log.Fatal(http.ListenAndServe(":9001", nil))
}
