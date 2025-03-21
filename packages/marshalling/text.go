package marshalling

import (
	"bytes"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	lib "github.com/vic/nix-versions/packages/versions"
)

func VersionsTable(versions []lib.Version, useColor bool) string {
	cyan := color.New(color.FgCyan)
	cyanS := cyan.SprintFunc()
	if !useColor {
		cyan.DisableColor()
	}

	var buff bytes.Buffer
	var tbl table.Table
	tbl = table.New("Name", "Attribute", cyanS("Version"), "Flake", "Revision").WithWriter(&buff).WithPrintHeaders(len(versions) > 1)
	for _, version := range versions {
		tbl = tbl.AddRow(version.Name, version.Attribute, cyanS(version.Version), version.Flake, version.Revision)
	}
	tbl.Print()
	return buff.String()
}
