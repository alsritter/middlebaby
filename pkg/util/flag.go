package util

import (
	"flag"
	"io/ioutil"

	"github.com/spf13/pflag"
)

type ignoredFlag struct {
	name string
}

func (ignoredFlag) String() string {
	return "ignored"
}

func (ignoredFlag) Type() string {
	return "string"
}

func (d ignoredFlag) Set(string) error {
	return nil
}

// IgnoredFlag ignores set value, without any warning
func IgnoredFlag(f *pflag.FlagSet, name, message string) {
	f.Var(ignoredFlag{name}, name, message)
}

// ParseConfigFileParameter Parse -config.file and -config.expand-env option via separate flag set, to avoid polluting default one and calling flag.Parse on it twice.
func ParseConfigFileParameter(args []string) (configFile string) {
	// ignore errors and any output here. Any flag errors will be reported by main flag.Parse() call.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)

	// usage not used in these functions.
	fs.StringVar(&configFile, "config.file", "", "")

	// Try to find -config.file and -config.expand-env option in the flags. As Parsing stops on the first error, eg. unknown flag, we simply
	// try remaining parameters until we find config flag, or there are no params left.
	// (ContinueOnError just means that flag.Parse doesn't call panic or os.Exit, but it returns error, which we ignore)
	for len(args) > 0 {
		_ = fs.Parse(args)
		args = args[1:]
	}

	return
}
