package versions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// TODO: Rename to Installable
// TODO: Produce as few different revisions as possible.
//
//	Each nixpkgs checkout is about 40M
//	--optimize=true
type Version struct {
	Name      string `json:"name"`
	Attribute string `json:"attr_path"`
	Version   string `json:"version"`
	Flake     string `json:"flake"`
	Revision  string `json:"revision"`
}

type ByVersion []Version

func (a ByVersion) Len() int      { return len(a) }
func (a ByVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByVersion) Less(i, j int) bool {
	x, err := semver.NewVersion(a[i].Version)
	if err != nil {
		return false
	}
	y, err := semver.NewVersion(a[j].Version)
	if err != nil {
		return false
	}
	return x.LessThan(y)
}

func SortByVersion(versions []Version) {
	sort.Sort(ByVersion(versions))
}

// Positive numbers take the last `n` items from list.
// Negative numbers tatke the first `n` items from list
func Limit(versions []Version, n int) []Version {
	if n > 0 {
		from := max(len(versions)-n, 0)
		to := min(len(versions)-1, (from + n))
		return versions[from : to+1]
	} else if n < 0 {
		n := n * -1
		to := min(n, len(versions))
		return versions[0:to]
	}
	return versions
}

func ConstraintBy(versions []Version, constraint string) ([]Version, error) {
	cond, err := semver.NewConstraint(constraint)
	if err != nil {
		return nil, fmt.Errorf("Could not create constraint from `%s`: %v\nSee https://github.com/Masterminds/semver?tab=readme-ov-file#basic-comparisons", constraint, err)
	}
	var res []Version
	for _, ver := range versions {
		v, err := semver.NewVersion(ver.Version)
		if err != nil {
			continue
		}
		if cond.Check(v) {
			res = append(res, ver)
		}
	}
	return res, nil
}

func Exact(versions []Version, attrPath string) []Version {
	var res []Version
	for _, ver := range versions {
		if ver.Attribute == attrPath {
			res = append(res, ver)
		}
	}
	return res
}

func VersionsJson(versions []Version) (string, error) {
	if len(versions) == 0 {
		return "", nil
	}
	var obj any
	if len(versions) == 1 {
		obj = versions[0]
	} else {
		obj = versions
	}
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func VersionsTable(versions []Version, useColor bool) string {
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

func Flake(versions []Version) (string, error) {
	jsonBytes, err := json.MarshalIndent(versions, "    ", "  ")
	if err != nil {
		return "", err
	}
	json := string(jsonBytes)

	var buff bytes.Buffer
	buff.WriteString("{\n")
	buff.WriteString("  inputs.nix-versions.url = \"github:vic/nix-versions\";\n")
	buff.WriteString("  inputs.nix-versions.inputs.nixpkgs.follows = \"nixpkgs\";\n")
	buff.WriteString("  inputs.nix-versions.inputs.treefmt-nix.inputs.nixpkgs.follows = \"nixpkgs\";\n")
	for _, version := range versions {
		name := version.Name
		attr := version.Attribute
		emptyRevs := []string{"master", "main", "HEAD", ""}
		var (
			url  string
			vers string
		)
		if slices.Contains(emptyRevs, version.Revision) {
			url = version.Flake
		} else {
			url = fmt.Sprintf("%s/%s", version.Flake, version.Revision)
		}
		vers = version.Version
		if version.Version == "" {
			vers = "latest"
		}
		comment := fmt.Sprintf(" # %s@%s", attr, vers)
		buff.WriteString(fmt.Sprintf("  inputs.\"%s\".url = \"%s\";%s\n", name, url, comment))
	}
	buff.WriteString("  outputs = inputs@{nixpkgs, self, ...}: inputs.nix-versions.lib.mkFlake {\n")
	buff.WriteString("    inherit inputs;\n")
	buff.WriteString(fmt.Sprintf("    nix-versions = builtins.fromJSON ''%s'';\n", json))
	buff.WriteString("  };\n")
	buff.WriteString("}\n")
	return buff.String(), nil
}

func Installables(versions []Version) string {
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
