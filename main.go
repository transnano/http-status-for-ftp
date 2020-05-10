// Copyright 2020 Transnano
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/kelseyhightower/envconfig"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "ftpchkr" // For Prometheus metrics.
)

var (
	listeningAddress = kingpin.Flag("address", "Address on which to expose metrics.").Default(":9065").String()
	metricsEndpoint  = kingpin.Flag("metrics", "Path under which to expose metrics.").Default("/metrics").String()
	healthEndpoint   = kingpin.Flag("health", "Path under which to check health.").Default("/status.html").String()
)

type Exporter struct {
	conf           FtpConfig
	mutex          sync.Mutex
	checkSuccesses prometheus.Counter
	checkFailures  prometheus.Counter
}

type FtpConfig struct {
	Host   string
	Port   string
	User   string
	Pass   string
	Origin string
}

func NewExporter() *Exporter {
	return &Exporter{
		checkSuccesses: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_successes_total",
			Help:      "Number of successes while health-checking ftp server.",
		}),
		checkFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while health-checking ftp server.",
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.checkSuccesses.Describe(ch)
	e.checkFailures.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping proftpd: %s", err)
		e.checkFailures.Inc()
		e.checkFailures.Collect(ch)
	}
}

func main() {

	// Parse flags
	log.AddFlags(kingpin.CommandLine)
	exporterName := namespace + "_exporter"
	kingpin.Version(version.Print(exporterName))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var fc FtpConfig
	if err := envconfig.Process("ftp", &fc); err != nil {
		log.Fatal(err.Error())
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err.Error())
	}
	fc.Origin = hostname

	exporter := NewExporter()
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector(exporterName))
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())

	log.Infoln("Starting "+exporterName, version.Info())
	log.Infoln("Build context", version.BuildContext())
	log.Infof("Starting Server: %s", *listeningAddress)

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	exporter.conf = fc
	server := newWebserver(exporter)
	go gracefullShutdown(server, quit, done)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	}

	<-done
	log.Infoln("Server stopped")
}

func gracefullShutdown(server *http.Server, quit <-chan os.Signal, done chan<- bool) {
	<-quit
	log.Infoln("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	close(done)
}

func newWebserver(e *Exporter) *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.HandleFunc(*healthEndpoint, func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusBadGateway
		c, err := ftp.Dial(e.conf.Host+e.conf.Port, ftp.DialWithTimeout(5*time.Second))
		if err != nil {
			log.Error(err)
		} else {
			if err = c.Login(e.conf.User, e.conf.Pass); err != nil {
				log.Error(err)
			} else {
				if err = c.Login(e.conf.User, e.conf.Pass); err != nil {
					log.Error(err)
				} else {
					if _, err := c.Retr("status_" + e.conf.Origin + ".html"); err != nil {
						log.Error(err)
					} else {
						if err := c.Quit(); err != nil {
							log.Error(err)
						} else {
							status = http.StatusOK
							e.checkSuccesses.Inc()
						}
					}
				}
			}
		}
		if status != http.StatusOK {
			e.checkFailures.Inc()
		}
		w.WriteHeader(status)
	})

	router.Handle(*metricsEndpoint, promhttp.Handler())

	return &http.Server{
		Handler:      router,
		ErrorLog:     log.NewErrorLogger(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}
