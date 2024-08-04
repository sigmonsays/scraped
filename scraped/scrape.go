package main

import "github.com/sigmonsays/scraped"

func Scrape(reg *scraped.Registry, pub *Publish) error {
	for _, p := range reg.Iterate() {
		err := p.Update()
		if err != nil {
			continue
		}
		buf, err := p.GetJson()
		if err != nil {
			continue
		}
		pt := p.GetType()
		log.Tracef("%s returned %s", pt.ID, buf)
		scraped.SetVar(pt.ID, p)

		// if configure, publish
		if pub != nil {
			pub.Publish(pt, buf)
		}
	}
	return nil
}
