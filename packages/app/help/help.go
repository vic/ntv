package help

import (
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

type HelpArgs struct {
	OnHelp func() `long:"help" short:"h"`
}

type CmdHelp struct {
	HelpTxt string
	HelpCtx func(cmd string) any
}

func (h *CmdHelp) ParseAndRun(args []string) ([]string, error) {
	cmd := args[0]
	a := HelpArgs{}
	a.OnHelp = func() {
		h.PrintHelpAndExit(cmd, 0)
	}
	parser := flags.NewParser(&a, flags.IgnoreUnknown)
	return parser.ParseArgs(args[1:])
}

func (h *CmdHelp) PrintHelpAndExit(cmd string, exitCode int) {
	t, err := template.New("HELP").Parse(h.HelpTxt)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	var out = os.Stdout
	if exitCode != 0 {
		out = os.Stderr
	}
	err = t.Execute(out, h.HelpCtx(cmd))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

type HelpDict map[string]CmdHelp

func (h HelpDict) PrintHelpAndExit(main CmdHelp, args []string, exitCode int) {
	var help = main
	var name = []string{}
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		name = append(name, arg)
		if x, ok := h[strings.Join(name, " ")]; ok {
			help = x
			break
		}
	}
	name = append([]string{args[0]}, name...)
	help.PrintHelpAndExit(strings.Join(name, " "), exitCode)
}
