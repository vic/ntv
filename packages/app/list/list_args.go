package list

import (
	_ "embed"
	"html/template"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-isatty"
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
	OnHelp         func() `long:"help" short:"h"`
	OnJSON         func() `long:"json" short:"j"`
	OnText         func() `long:"text" short:"t"`
	OnInstallable  func() `long:"installable" short:"i"`
	OnFlake        func() `long:"flake" short:"f"`
	OnAll          func() `long:"all" short:"a"`
	OnOne          func() `long:"one" short:"1"`
	OnNixHub       func() `long:"nixhub" short:"n"`
	OnLazamar      func() `long:"lazamar" short:"l"`
	OutFmt         OutFmt
	ShowOpt        ShowOpt
	LazamarChannel *string `long:"channel" short:"c"`
	Color          bool    `long:"color" short:"C"`
	rest           []string
}

//go:embed HELP
var HELP string

func HelpAndExit(name string) {
	t, err := template.New("HELP").Parse(HELP)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err = t.Execute(os.Stdout, map[string]interface{}{
		"Cmd": name,
	})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func NewListArgs() *ListArgs {
	args := ListArgs{
		OutFmt:  OutText,
		ShowOpt: ShowConstrained,
		Color:   isatty.IsTerminal(os.Stdout.Fd()),
	}
	args.OnHelp = func() {
		HelpAndExit("nix-versions")
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
