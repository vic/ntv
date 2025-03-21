package versions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/rodaine/table"
)

// TODO: Rename to Installable
// TODO: Produce as few different revisions as possible.
//
//	Each nixpkgs checkout is about 40M
//	--optimize=true
type Version struct {
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

func VersionsTable(versions []Version) string {
	var buff bytes.Buffer
	var tbl table.Table
	tbl = table.New("Version", "Attribute", "Flake", "Revision").WithWriter(&buff).WithPrintHeaders(len(versions) > 1)
	for _, version := range versions {
		tbl = tbl.AddRow(version.Version, version.Attribute, version.Flake, version.Revision)
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
		name := version.Attribute
		emptyRevs := []string{"master", "main", "HEAD", ""}
		var (
			url     string
			comment string
		)
		if slices.Contains(emptyRevs, version.Revision) {
			url = version.Flake
		} else {
			url = fmt.Sprintf("%s/%s", version.Flake, version.Revision)
		}
		if version.Version != "" {
			comment = fmt.Sprintf(" # version: %s", version.Version)
		}
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
	// if piped {
	// 	for n, version := range versions {
	// 		buff.WriteString(fmt.Sprintf("%s/%s#%s", version.Flake, version.Revision, version.Attribute))
	// 		if n < len(versions)-1 {
	// 			buff.WriteString("\n")
	// 		}
	// 	}
	// 	return buff.String()
	// }
	var tbl table.Table
	tbl = table.New("Installable", "Comment").WithWriter(&buff).WithPrintHeaders(false)
	for _, version := range versions {
		installable := fmt.Sprintf("%s/%s#%s", version.Flake, version.Revision, version.Attribute)
		var comment string
		if version.Version != "" {
			comment = fmt.Sprintf("# version: %s", version.Version)
		}
		tbl = tbl.AddRow(installable, comment)
	}
	tbl.Print()
	return buff.String()
}
