package marshalling

import (
	"bytes"
	"fmt"

	"github.com/rodaine/table"
	lib "github.com/vic/nix-versions/packages/versions"
)

func Installables(versions []lib.Version) string {
	var buff bytes.Buffer
	// stat, _ := os.Stdout.Stat()
	// piped := stat.Mode()&os.ModeCharDevice == 0
	var tbl table.Table
	tbl = table.New("Installable", "Comment").WithWriter(&buff).WithPrintHeaders(false)
	for _, version := range versions {
		installable := fmt.Sprintf("%s/%s#%s", version.Flake, version.Revision, version.Attribute)
		var vrs string = version.Version
		if vrs == "" {
			vrs = "latest"
		}
		comment := fmt.Sprintf("# %s@%s", version.Name, vrs)
		tbl = tbl.AddRow(installable, comment)
	}
	tbl.Print()
	return buff.String()
}
