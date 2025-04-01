package list

import (
	_ "embed"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-isatty"
	"github.com/vic/ntv/packages/app/help"
	"github.com/vic/ntv/packages/search_spec"
)

type OutFmt uint8

const (
	OutJSON OutFmt = iota
	OutText
	OutInstallable
	OutFlake
)

type ShowOpt uint8

const (
	ShowAll ShowOpt = iota
	ShowOne
	ShowConstrained
)

type ListArgs struct {
	OnJSON           func()       `long:"json" short:"j"`
	OnText           func()       `long:"text" short:"t"`
	OnInstallable    func()       `long:"installable" short:"i"`
	OnFlake          func()       `long:"flake" short:"f"`
	OnAll            func()       `long:"all" short:"a"`
	OnOne            func()       `long:"one" short:"1"`
	OnNixHub         func()       `long:"nixhub"`
	OnLazamar        func()       `long:"lazamar"`
	OnNixPackagesCom func()       `long:"history"`
	OnRead           func(string) `long:"read" short:"r"`
	ReadFiles        []string
	OutFmt           OutFmt
	ShowOpt          ShowOpt
	OnLazamarChannel func(string) `long:"channel"`
	Color            bool         `long:"color" short:"C"`
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

func NewListArgs() *ListArgs {
	args := ListArgs{
		OutFmt:          OutText,
		ShowOpt:         ShowConstrained,
		Color:           isatty.IsTerminal(os.Stdout.Fd()),
		ReadFiles:       []string{},
		versionsBackend: search_spec.VersionsBackend{NixHub: &search_spec.Unit{}},
	}
	args.OnRead = func(file string) {
		args.ReadFiles = append(args.ReadFiles, file)
	}
	args.OnJSON = func() {
		args.OutFmt = OutJSON
	}
	args.OnText = func() {
		args.OutFmt = OutText
	}
	args.OnInstallable = func() {
		args.OutFmt = OutInstallable
	}
	args.OnFlake = func() {
		args.OutFmt = OutFlake
	}
	args.OnAll = func() {
		args.ShowOpt = ShowAll
	}
	args.OnOne = func() {
		args.ShowOpt = ShowOne
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

func (a *ListArgs) Parse(args []string) error {
	parser := flags.NewParser(a, flags.AllowBoolValues|flags.IgnoreUnknown)
	rest, err := parser.ParseArgs(args)
	if err != nil {
		return err
	}
	a.rest = rest
	return nil
}

func (a *ListArgs) ParseAndRun(args []string) error {
	err := a.Parse(args)
	if err != nil {
		return err
	}
	return a.Run()
}
