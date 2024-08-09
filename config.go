package scraped

import (
	"bytes"
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// main configuration structure
type AppConfig struct {
	ReloadInterval int       `yaml:"reload_interval"`
	ScrapeInterval int       `yaml:"scrape_interval"`
	HTTP           *HTTP     `yaml:"http"`
	ScriptDir      string    `yaml:"script_dir"`
	Plugins        []*Plugin `yaml:"plugins"`
	Hostname       string    `yaml:"hostname"`
	PublishUrl     string    `yaml:"publish_url"`
}

type HTTP struct {
	Addr string `yaml:"addr"`
}

type Plugin struct {
	ID     string    `yaml:"id"`
	Type   string    `yaml:"type"`
	Params yaml.Node `yaml:"params"`
}

func (c *AppConfig) LoadYaml(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(nil)
	_, err = b.ReadFrom(f)
	if err != nil {
		return err
	}

	if err := c.LoadYamlBuffer(b.Bytes()); err != nil {
		return err
	}

	if err := c.FixupConfig(); err != nil {
		return err
	}

	return nil
}

func (c *AppConfig) LoadYamlBuffer(buf []byte) error {
	err := yaml.Unmarshal(buf, c)
	if err != nil {
		return err
	}
	return nil
}

func (me *AppConfig) PrintConfig() {
	d, err := yaml.Marshal(me)
	if err != nil {
		fmt.Println("Marshal error", err)
		return
	}
	fmt.Println("-- Configuration --")
	fmt.Println(string(d))
}

func GetDefaultConfig() *AppConfig {
	cfg := &AppConfig{}
	cfg.ReloadInterval = 3600
	cfg.ScrapeInterval = 120
	cfg.HTTP = &HTTP{}
	cfg.HTTP.Addr = ":5959"
	return cfg
}

func (c *AppConfig) LoadDefault() {
	*c = *GetDefaultConfig()
}

// after loading configuration this gives us a spot to "fix up" any configuration
// or abort the loading process
func (c *AppConfig) FixupConfig() error {
	// var emptyConfig AppConfig

	return nil
}

func PrintDefaultConfig() {
	conf := GetDefaultConfig()
	conf.PrintConfig()
}
