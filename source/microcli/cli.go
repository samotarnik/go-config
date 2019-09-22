package microcli

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/micro/cli"
	"github.com/samotarnik/go-config/source"
	"github.com/micro/go-micro/cmd"
)

type cliSource struct {
	opts source.Options
	ctx  *cli.Context
}

func (c *cliSource) Read() (*source.ChangeSet, error) {
	var changes map[string]interface{}

	for _, name := range c.ctx.GlobalFlagNames() {
		tmp := toEntry(name, c.ctx.GlobalGeneric(name))
		mergo.Map(&changes, tmp) // need to sort error handling
	}

	for _, name := range c.ctx.FlagNames() {
		tmp := toEntry(name, c.ctx.Generic(name))
		mergo.Map(&changes, tmp) // need to sort error handling
	}

	b, err := c.opts.Encoder.Encode(changes)
	if err != nil {
		return nil, err
	}

	cs := &source.ChangeSet{
		Format:    c.opts.Encoder.String(),
		Data:      b,
		Timestamp: time.Now(),
		Source:    c.String(),
	}
	cs.Checksum = cs.Sum()

	return cs, nil
}

func toEntry(name string, v interface{}) map[string]interface{} {
	n := strings.ToLower(name)
	keys := strings.FieldsFunc(n, split)
	reverse(keys)
	tmp := make(map[string]interface{})
	for i, k := range keys {
		if i == 0 {
			tmp[k] = v
			continue
		}

		tmp = map[string]interface{}{k: tmp}
	}
	return tmp
}

func reverse(ss []string) {
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}
}

func split(r rune) bool {
	return r == '-' || r == '_'
}

func (c *cliSource) Watch() (source.Watcher, error) {
	return source.NewNoopWatcher()
}

func (c *cliSource) String() string {
	return "microcli"
}

// NewSource returns a config source for integrating parsed flags from a micro/cli.Context.
// Hyphens are delimiters for nesting, and all keys are lowercased. The assumption is that
// command line flags have already been parsed.
//
// Example:
//      cli.StringFlag{Name: "db-host"},
//
//
//      {
//          "database": {
//              "host": "localhost"
//          }
//      }
func NewSource(opts ...source.Option) source.Source {
	options := source.NewOptions(opts...)

	var ctx *cli.Context

	c, ok := options.Context.Value(contextKey{}).(*cli.Context)
	if ok {
		ctx = c
	}

	// no context
	if ctx == nil {
		// get the default app/flags
		app := cmd.App()
		flags := app.Flags

		// create flagset
		set := flag.NewFlagSet(app.Name, flag.ContinueOnError)

		// apply flags to set
		for _, f := range flags {
			f.Apply(set)
		}

		// parse flags
		set.SetOutput(ioutil.Discard)
		set.Parse(os.Args[1:])

		// normalise flags
		normalizeFlags(app.Flags, set)

		// create context
		ctx = cli.NewContext(app, set, nil)
	}

	return &cliSource{
		ctx:  ctx,
		opts: options,
	}
}

// WithContext returns a new source with the context specified.
// The assumption is that Context is retrieved within an app.Action function.
func WithContext(ctx *cli.Context, opts ...source.Option) source.Source {
	return &cliSource{
		ctx:  ctx,
		opts: source.NewOptions(opts...),
	}
}
