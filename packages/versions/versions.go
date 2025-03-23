package versions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
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
	constraint = strings.Replace(constraint, "latest", "", 1)
	if strings.TrimSpace(constraint) == "" {
		constraint = "*"
	}
	cond, err := semver.NewConstraint(constraint)
	if err != nil {
		return nil, fmt.Errorf("could not create constraint from string `%s`:\n%v\nSee https://github.com/Masterminds/semver?tab=readme-ov-file#basic-comparisons", constraint, err)
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
