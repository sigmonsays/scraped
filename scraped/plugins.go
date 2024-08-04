package main

import (
	"os"

	"github.com/sigmonsays/scraped"
)

func LoadPlugins(cfg *scraped.AppConfig, reg *scraped.Registry) error {
	reg.UnloadAll()
	os.MkdirAll(cfg.ScriptDir, 0755)
	for _, pcfg := range cfg.Plugins {
		if pcfg.Type == "exec" {
			p := scraped.DefaultExec()
			p.ID = pcfg.ID
			err := pcfg.Params.Decode(&p)
			if err != nil {
				log.Warnf("Decode %s", err)
			}
			p.Init(cfg.ScriptDir, pcfg.ID)
			reg.Register(p)
		} else {
			log.Warnf("Unknown scraper type %s", pcfg.Type)
		}
	}
	return nil
}
