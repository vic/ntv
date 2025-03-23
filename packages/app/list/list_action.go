package list

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/vic/ntv/packages/app/new"
	"github.com/vic/ntv/packages/flake"
	"github.com/vic/ntv/packages/search"
	"github.com/vic/ntv/packages/search_spec"
)

func (a *ListArgs) Run() error {
	specs, err := search_spec.ParseSearchSpecs(a.rest, a.LazamarChannel)
	if err != nil {
		return err
	}

	res, err := search.PackageSearchSpecs(specs).Search()
	if err != nil {
		return err
	}

	var out string
	if a.OutFmt == OutText {
		out, err = a.TextOut(res)
		if err != nil {
			return err
		}
	}

	if a.OutFmt == OutInstallable {
		out, err = InstallableOut(res)
		if err != nil {
			return err
		}
	}

	if a.OutFmt == OutJSON {
		out, err = JsonOut(res)
		if err != nil {
			return err
		}
	}

	if a.OutFmt == OutFlake {
		f := flake.New()
		out, err = new.FlakeCode(f, res)
		if err != nil {
			return err
		}
	}

	fmt.Println(out)
	return nil
}

func JsonOut(res search.PackageSearchResults) (string, error) {
	if err := res.EnsureOneSelected(); err != nil {
		return "", err
	}
	if err := res.EnsureUniquePackageNames(); err != nil {
		return "", err
	}

	var tools = make([]flake.Tool, 0)

	for _, r := range res {
		tools = append(tools, flake.AsTool(r))
	}

	jsonBytes, err := json.MarshalIndent(&tools, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func InstallableOut(res search.PackageSearchResults) (string, error) {
	if err := res.EnsureOneSelected(); err != nil {
		return "", err
	}
	if err := res.EnsureUniquePackageNames(); err != nil {
		return "", err
	}

	buff := bytes.Buffer{}
	for _, r := range res {
		fmt.Fprintf(&buff, "%s\n", r.Installable())
	}
	return buff.String(), nil
}

func (a *ListArgs) TextOut(res search.PackageSearchResults) (string, error) {
	color.NoColor = !a.Color

	nocolor := color.New(color.FgWhite)
	constrainedColor := color.New(color.FgCyan)
	selectedColor := color.New(color.FgHiGreen)

	hd := nocolor.SprintfFunc()

	buff := bytes.Buffer{}
	tbl := table.New(hd("Name"), hd("Attribute"), hd("Version"), hd("Flake"), hd("Revision")).WithWriter(&buff)

	for _, r := range res {
		for _, v := range r.Versions {
			var otherColor = nocolor.SprintfFunc()
			var versionColor = nocolor.SprintfFunc()

			isSelected := r.Selected == v
			isConstrained := slices.Contains(r.Constrained, v)

			if isConstrained {
				versionColor = constrainedColor.SprintfFunc()
			}

			if isSelected {
				versionColor = selectedColor.SprintfFunc()
				otherColor = color.New(color.FgHiWhite).SprintfFunc()
			}

			if a.ShowOpt == ShowOne && !isSelected {
				continue
			}

			if a.ShowOpt == ShowConstrained && !isConstrained && len(r.Constrained) > 0 {
				continue
			}

			tbl.AddRow(
				otherColor(v.Name),
				otherColor(v.Attribute),
				versionColor(v.Version),
				otherColor(v.Flake),
				otherColor(v.Revision),
			)
		}
	}

	tbl.Print()
	return buff.String(), nil
}
