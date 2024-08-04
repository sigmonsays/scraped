package scraped

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	gologging "github.com/sigmonsays/go-logging"
)

func DefaultExec() *Exec {
	l := gologging.Register("exec", func(newlog gologging.Logger) { log = newlog })
	e := &Exec{
		log:        l,
		TrimResult: 512,
	}
	return e
}

func SplitStringWithQuotes(s string) []string {
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)

		}
	}
	m := strings.FieldsFunc(s, f)
	return m
}

type Exec struct {
	log         gologging.Logger
	ID          string   `json:"id"`
	Debug       bool     `json:"debug,omitempty"`
	Command     []string `json:"command"`
	Timeout     int      `json:"timeout"`
	Interpreter string   `json:"interpreter"`
	Shell       bool     `json:"shell"`
	Script      string   `json:"script"`
	TrimResult  int      `json:"trim_result"`
	Env         []string `json:"env,omitempty"`
	ShortOutput bool     `json:"short_output,omitempty"`
	DoChecksum  bool     `json:"do_checksum,omitempty"`
	// plugin spits out json
	Json         bool `json:"json,omitempty"`
	BinaryOutput bool `json:"binary_output,omitempty"`
	NoStore      bool `json:"nostore,omitempty"`

	// results
	ExitCode      int         `json:"exit_code"`
	Result        interface{} `json:"result"`
	TimedOut      bool        `json:"timed_out"`
	ExecuteTimeMs int64       `json:"execute_time_ms"`
	Checksum      string      `json:"checksum,omitempty"`
}

func (me *Exec) Init(scriptdir string, id string) error {

	if me.Shell && me.Script != "" {
		if me.Interpreter == "" {
			me.Interpreter = "/bin/sh"
		}

		scriptfile := filepath.Join(scriptdir, id+".sh")
		me.Command = []string{
			me.Interpreter,
			"-c",
			scriptfile,
		}
		me.Shell = false
		me.log.Tracef("switching command to %v", me.Command)
		me.log.Tracef("writing script %s", scriptfile)
		ioutil.WriteFile(scriptfile, []byte(me.Script), 0755)

	}
	// me.Command is assumed to be set
	return nil
}

func (me *Exec) Validate() error {
	if len(me.Command) == 0 {
		return fmt.Errorf("command is missing")
	}
	return nil
}

func (me *Exec) String() string {
	buf, err := me.GetJson()
	if err != nil {
		me.log.Warnf("GetJson Exec %s", err)
		return ""
	}
	return string(buf)
}

func NewExecDataItem(cmdline []string) (Scraper, error) {
	return NewExec(cmdline)
}

func NewExec(cmdline []string) (*Exec, error) {
	e := &Exec{
		Command: cmdline,
	}
	return e, nil
}

func (me *Exec) GetType() *PluginType {
	ret := &PluginType{
		Type: "exec",
		ID:   me.ID,
	}
	return ret
}

func (me *Exec) buildCommand(ctx context.Context) (*exec.Cmd, func()) {
	var cancel func()

	if me.Timeout > 0 {
		ctx2, cancel2 := context.WithTimeout(ctx, time.Duration(me.Timeout)*time.Second)
		cancel = func() {
			if cancel2 == nil {
				me.log.Tracef("no cancel func for exec context")
			} else {
				me.log.Tracef("cancelling exec context")
				cancel2()
			}
		}
		ctx = ctx2
	}

	cmdline := []string{}

	if me.Shell {
		cmdline = []string{"/bin/sh", "-c"}
		cmdline = append(cmdline, strings.Join(me.Command, " "))

	} else {
		cmdline = me.Command
	}

	me.log.Tracef("execute command %#v (timeout:%d trim-result:%d)", cmdline, me.Timeout, me.TrimResult)

	c := exec.CommandContext(ctx, cmdline[0], cmdline[1:]...)
	c.Stderr = os.Stderr
	env := make([]string, 0)
	env = append(env, os.Environ()...)
	env = append(env, me.Env...)
	c.Env = env
	// me.log.Tracef("setting %d environment variables: %v", len(env), env)

	// always keep cancel func callable
	if cancel == nil {
		cancel = func() {}
	}
	return c, cancel
}

func (me *Exec) Update() (err error) {
	me.log.Tracef("exec update")
	now := time.Now()

	ctx := context.Background()
	c, cancel := me.buildCommand(ctx)
	defer cancel()

	buf := bytes.NewBuffer(nil)
	c.Stdout = buf

	err = c.Run()

	dur := time.Since(now)
	me.ExecuteTimeMs = int64(dur.Nanoseconds() / 1000000)

	if ctx.Err() == context.DeadlineExceeded {
		me.TimedOut = true
	}

	if err != nil {

		if c.ProcessState != nil {
			me.ExitCode = c.ProcessState.ExitCode()
			me.Result = ""
		}
		return err
	}
	out := strings.TrimRight(buf.String(), "\n")

	me.ExitCode = c.ProcessState.ExitCode()

	if me.DoChecksum {
		cs := sha256.New()
		cs.Write([]byte(out))
		checksum := cs.Sum(nil)
		me.Checksum = base64.URLEncoding.EncodeToString(checksum)
	}

	if me.NoStore {
		me.log.Trace("result of exec is not stored/cached")
	} else {
		if me.Json {
			me.Result = json.RawMessage([]byte(out))

		} else if me.BinaryOutput {
			me.Result = buf.Bytes()

		} else {
			me.Result = out
		}
	}

	me.log.Tracef("return err %s", err)
	return err
}

func (me *Exec) ApiEndpoint(fqdn string) string {
	return ""
}

func (me *Exec) GetBinary() ([]byte, error) {
	if me.BinaryOutput == false {
		return nil, fmt.Errorf("BinaryOutput not enabled")
	}
	me.log.Tracef("GetBinary %T", me.Result)
	buf, ok := me.Result.([]byte)
	if ok == false {
		return nil, fmt.Errorf("Result is not bytes")
	}
	return buf, nil
}

func (me *Exec) GetJson() ([]byte, error) {
	if me.ShortOutput {
		return json.Marshal(me.Result)
	}

	return json.Marshal(me)
}

func (me *Exec) IsEmpty() bool {
	buf, err := json.Marshal(me.Result)
	if err == nil {

		// empty string
		if bytes.Compare(buf, []byte(`""`)) == 0 {
			return true
		}

	}
	return false
}
