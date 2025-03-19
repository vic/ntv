package versions

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/rodaine/table"
)

type Version struct {
	Attribute string `json:"attr_path"`
	Version   string `json:"version"`
	Revision  string `json:"nixpkgs_rev"`
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
		return nil, err
	}
	var res []Version
	for _, ver := range versions {
		v, err := semver.NewVersion(ver.Version)
		if err != nil {
			return nil, err
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
	tbl = table.New("Version", "Attribute", "Nixpkgs-Revision").WithWriter(&buff).WithPrintHeaders(len(versions) > 1)
	for _, version := range versions {
		tbl = tbl.AddRow(version.Version, version.Attribute, version.Revision)
	}
	tbl.Print()
	return buff.String()
}
