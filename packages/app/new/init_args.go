package new

import (
	_ "embed"

	"github.com/jessevdk/go-flags"
	"github.com/vic/ntv/packages/app/help"
	"github.com/vic/ntv/packages/search_spec"
)

type InitArgs struct {
	OnNixHub         func()       `long:"nixhub" short:"n"`
	OnLazamar        func()       `long:"lazamar" short:"l"`
	OnLazamarChannel func(string) `long:"channel" short:"c"`
	OnNixPackagesCom func()       `long:"history" short:"h"`
	NtvFlake         string       `long:"override-ntv"`
	versionsBackend  search_spec.VersionsBackend
	rest             []string
}

//go:embed HELP
var HELP string

var Help = help.CmdHelp{
	HelpTxt: HELP,
	HelpCtx: func(name string) any {
		return map[string]interface{}{
			"Cmd": name,
		}
	},
}

func NewInitArgs() *InitArgs {
	args := InitArgs{
		versionsBackend: search_spec.VersionsBackend{
			NixHub: &search_spec.Unit{},
		},
	}

	args.OnNixHub = func() {
		args.versionsBackend = search_spec.VersionsBackend{NixHub: &search_spec.Unit{}}
	}
	args.OnLazamar = func() {
		if args.versionsBackend.LazamarChannel != nil {
			return
		}
		args.OnLazamarChannel("nixpkgs-unstable")
	}
	args.OnLazamarChannel = func(channel string) {
		args.versionsBackend = search_spec.VersionsBackend{LazamarChannel: (*search_spec.LazamarChannel)(&channel)}
	}
	args.OnNixPackagesCom = func() {
		args.versionsBackend = search_spec.VersionsBackend{NixPackagesCom: &search_spec.Unit{}}
	}
	return &args
}

func (a *InitArgs) Parse(args []string) error {
	parser := flags.NewParser(a, flags.AllowBoolValues|flags.IgnoreUnknown)
	rest, err := parser.ParseArgs(args)
	if err != nil {
		return err
	}
	a.rest = rest
	return nil
}

func (a *InitArgs) ParseAndRun(args []string) error {
	err := a.Parse(args)
	if err != nil {
		return err
	}
	return a.Run()
}
