package main

import (
	"expvar"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/sigmonsays/scraped"

	gologging "github.com/sigmonsays/go-logging"
)

func main() {
	loglevel := "info"
	configfile := "/etc/whatever.yaml"
	flag.StringVar(&configfile, "config", configfile, "specify config file")
	flag.StringVar(&loglevel, "loglevel", loglevel, "log level")
	flag.Parse()

	gologging.SetLogLevel(loglevel)

	cfg, err := LoadConfig(configfile)
	if err != nil {
		ExitError("LoadYaml %s: %s", configfile, err)
	}

	storeHandler := &scraped.Store{}

	mx := http.NewServeMux()
	mx.HandleFunc("/vars", expvar.Handler().ServeHTTP)
	mx.HandleFunc("/store", storeHandler.StoreHandler)
	srv := &http.Server{}
	srv.Handler = mx
	srv.Addr = cfg.HTTP.Addr
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Warnf("ListenAndServe %s: %s", cfg.HTTP.Addr)
		}
	}()

	control := make(chan int, 5)
	control <- 1 // reload config
	control <- 2 // trigger plugins

	reloadTicker := time.NewTicker(time.Duration(cfg.ReloadInterval) * time.Second)
	defer reloadTicker.Stop()

	scrapeTicker := time.NewTicker(time.Duration(cfg.ScrapeInterval) * time.Second)
	defer scrapeTicker.Stop()

	reg := scraped.NewRegistry()

	pub := NewPublish(cfg.Hostname, cfg.PublishUrl)

Loop:
	for {
		select {
		case <-scrapeTicker.C:
			Scrape(reg, pub)

		case <-reloadTicker.C:
			control <- 1

		case code := <-control:

			if code == 0 { // Exit program
				break Loop

			} else if code == 1 { // reload config
				newcfg, err := LoadConfig(configfile)
				if err == nil {
					cfg = newcfg
				} else {
					log.Warnf("LoadConfig %s: %s", configfile, err)
				}
				scrapeTicker.Reset(time.Duration(cfg.ScrapeInterval) * time.Second)

				LoadPlugins(cfg, reg)
				pub = NewPublish(cfg.Hostname, cfg.PublishUrl)

			} else if code == 2 { // scrape
				Scrape(reg, pub)

			} else {
				log.Debugf("Unknown control code: %", code)
			}
		}
	}
}

func LoadConfig(configfile string) (*scraped.AppConfig, error) {
	cfg := scraped.GetDefaultConfig()
	exists := false
	st, err := os.Stat(configfile)
	if err == nil && st.IsDir() == false {
		exists = true
	}
	if !exists {
		return cfg, nil
	}
	err = cfg.LoadYaml(configfile)
	if err != nil {
		return cfg, err
	}

	if cfg.Hostname == "" {
		h, err := os.Hostname()
		if err == nil {
			cfg.Hostname = h
		}
	}

	return cfg, nil
}
