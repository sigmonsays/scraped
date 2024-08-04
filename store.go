package scraped

import (
	"encoding/json"
	"expvar"
	"io"
	"net/http"

	gologging "github.com/sigmonsays/go-logging"
)

var (
	stored_vars *expvar.Map
)

func init() {
	stored_vars = expvar.NewMap("stored_values")
}

func SetStoredVar(hostname, name string, value json.RawMessage) {
	x := stored_vars.Get(hostname)
	if x == nil {
		x = &expvar.Map{}
		stored_vars.Set(hostname, x)
	}
	x2, _ := x.(*expvar.Map)
	v := NewJsonMessage(value)
	x2.Set(name, v)
	log.Debugf("stored value for hostname %s, name %s, %d bytes", hostname, name, len(value))
}

type Store struct {
	log gologging.Logger
}

func (me *Store) StoreHandler(w http.ResponseWriter, r *http.Request) {

	// read body
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		me.log.Warnf("ReadAll %s", err)
		return
	}

	// decode request
	var v *PublishedValue
	err = json.Unmarshal(buf, &v)
	if err != nil {
		me.log.Warnf("Unmarshal %s", err)
		return
	}

	// store request
	SetStoredVar(v.Hostname, v.ID, v.Value)
}
