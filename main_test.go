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
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewExporter(t *testing.T) {
	namespace := "ftpchkr"
	tests := []struct {
		name string
		want *Exporter
	}{
		// TODO: Add test cases.
		{
			"Successful",
			&Exporter{
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewExporter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewExporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExporter_Describe(t *testing.T) {
	type fields struct {
		conf           FtpConfig
		mutex          sync.Mutex
		checkSuccesses prometheus.Counter
		checkFailures  prometheus.Counter
	}
	type args struct {
		ch chan<- *prometheus.Desc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Exporter{
				conf:           tt.fields.conf,
				mutex:          tt.fields.mutex,
				checkSuccesses: tt.fields.checkSuccesses,
				checkFailures:  tt.fields.checkFailures,
			}
			e.Describe(tt.args.ch)
		})
	}
}

func TestExporter_Collect(t *testing.T) {
	type fields struct {
		conf           FtpConfig
		mutex          sync.Mutex
		checkSuccesses prometheus.Counter
		checkFailures  prometheus.Counter
	}
	type args struct {
		ch chan<- prometheus.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Exporter{
				conf:           tt.fields.conf,
				mutex:          tt.fields.mutex,
				checkSuccesses: tt.fields.checkSuccesses,
				checkFailures:  tt.fields.checkFailures,
			}
			e.Collect(tt.args.ch)
		})
	}
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_gracefullShutdown(t *testing.T) {
	type args struct {
		server *http.Server
		quit   <-chan os.Signal
		done   chan<- bool
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gracefullShutdown(tt.args.server, tt.args.quit, tt.args.done)
		})
	}
}

func Test_newWebserver(t *testing.T) {
	type args struct {
		e *Exporter
	}
	tests := []struct {
		name string
		args args
		want *http.Server
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newWebserver(tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newWebserver() = %v, want %v", got, tt.want)
			}
		})
	}
}
