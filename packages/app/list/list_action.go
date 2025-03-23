package list

import (
	"bytes"
	"fmt"
	"os/exec"
	"slices"

	"github.com/fatih/color"
	"github.com/rodaine/table"

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
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%v: %s", err, string(ee.Stderr))
		}
		return err
	}

	if len(res) == 0 {
		return fmt.Errorf("no results found")
	}

	var out string
	if a.OutFmt == OutText {
		out, err = a.Text(res)
		if err != nil {
			return err
		}
	}

	if a.OutFmt == OutInstallable {
		out, err = a.Installable(res)
		if err != nil {
			return err
		}
	}

	fmt.Println(out)
	return nil
}

func (a *ListArgs) Installable(res search.PackageSearchResults) (string, error) {
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

func (a *ListArgs) Text(res search.PackageSearchResults) (string, error) {
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
