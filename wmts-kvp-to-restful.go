package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PDOK/wmts-kvp-to-restful/operations"
	"github.com/go-chi/chi"
)

const (
	shutdownTimeout = 15 * time.Second
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
	shutdownDelay := flag.Int("d", 0, "Delay (in seconds) before initiating graceful shutdown (e.g. useful in k8s to allow ingress controller to update their endpoints list, default: 0")
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

	router := chi.NewRouter()
	proxy := &httputil.ReverseProxy{Director: director}

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(`{"health": "OK"}`))
		return
	})

	log.Println("wmts-kvp-to-restful started")

	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {

		// Simple logging ...
		// TODO put logging on a chan for async output
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

	err := startServer("wmts-kvp-to-restful", ":9001", *shutdownDelay, router)
	if err != nil {
		log.Fatal(err)
	}
}

// startServer creates and starts an HTTP server, also takes care of graceful shutdown
func startServer(name string, address string, shutdownDelay int, router http.Handler) error {
	// Create HTTP server
	server := http.Server{
		Addr:    address,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go func() {
		log.Printf("%s listening on %s", name, address)
		// ListenAndServe always returns a non-nil error. After Shutdown or
		// Close, the returned error is ErrServerClosed
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to shutdown %s: %v", name, err)
		}
	}()

	// Listen for interrupt signal and then perform shutdown
	<-ctx.Done()
	stop()

	if shutdownDelay > 0 {
		log.Printf("stop signal received, initiating shutdown of %s after %d seconds delay", name, shutdownDelay)
		time.Sleep(time.Duration(shutdownDelay) * time.Second)
	}
	log.Printf("shutting down %s gracefully", name)

	// Shutdown with a max timeout.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return server.Shutdown(timeoutCtx)
}
