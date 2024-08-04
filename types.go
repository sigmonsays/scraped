package scraped

import "encoding/json"

type Scraper interface {

	// used value is exported to expvars
	String() string

	// get the data item value
	GetJson() ([]byte, error)

	// get the data item type
	GetType() *PluginType

	// refresh data
	Update() error
}

type PluginType struct {
	ID   string
	Type string
}

type PublishedValue struct {
	Hostname string
	ID       string
	Type     string
	Value    json.RawMessage
}
