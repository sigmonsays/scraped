package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sigmonsays/scraped"
)

func NewPublish(hostname, url string) *Publish {
	return &Publish{
		Hostname: hostname,
		Url:      url,
	}
}

type Publish struct {
	Hostname string
	Url      string
}

func (me *Publish) Publish(pt *scraped.PluginType, buf []byte) error {
	pv := &scraped.PublishedValue{
		ID:       pt.ID,
		Type:     pt.Type,
		Hostname: me.Hostname,
		Value:    json.RawMessage(buf),
	}
	buf2, err := json.Marshal(pv)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(buf2)
	req, err := http.NewRequest("POST", me.Url, body)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	code := res.StatusCode
	if code != 200 {
		return fmt.Errorf("status-code %d", code)
	}

	log.Debugf("Published %s to %s", pt.ID, me.Url)

	return nil
}
