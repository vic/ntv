package new

import (
	_ "embed"

	"github.com/jessevdk/go-flags"
	"github.com/vic/ntv/packages/app/help"
)

type InitArgs struct {
	OnNixHub       func()  `long:"nixhub" short:"n"`
	OnLazamar      func()  `long:"lazamar" short:"l"`
	LazamarChannel *string `long:"channel" short:"c"`
	NtvFlake       string  `long:"override-ntv"`
	rest           []string
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
	args := InitArgs{}
	args.OnNixHub = func() {
		args.LazamarChannel = nil
	}
	args.OnLazamar = func() {
		if args.LazamarChannel == nil {
			channel := "nixpkgs-unstable"
			args.LazamarChannel = &channel
		}
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
