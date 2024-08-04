package scraped

import (
	"encoding/json"
	"expvar"
)

var (
	vars *expvar.Map
)

func init() {
	vars = expvar.NewMap("values")
}

func SetVar(name string, value interface{}) {
	v := NewJsonMessage(value)
	vars.Set(name, v)
}

func NewJsonMessage(v interface{}) *JsonMessage {
	return &JsonMessage{v}
}

type JsonMessage struct {
	value interface{}
}

func (me *JsonMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(me.value)
}

func (me *JsonMessage) String() string {
	buf, _ := json.Marshal(me)
	return string(buf)
}

func (me *JsonMessage) Zero() {
	me.value = nil
}
