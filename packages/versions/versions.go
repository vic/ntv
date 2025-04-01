package versions

import (
	"fmt"
	"regexp"
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

type ByVersion []*Version

func (a ByVersion) Len() int      { return len(a) }
func (a ByVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByVersion) Less(i, j int) bool {
	xVer := a[i].Version
	yVer := a[j].Version
	if !strings.Contains(xVer, ".") {
		return true
	}
	if !strings.Contains(yVer, ".") {
		return false
	}
	x, errx := semver.NewVersion(xVer)
	y, erry := semver.NewVersion(yVer)
	if errx != nil {
		return true
	}
	if erry != nil {
		return false
	}
	return x.LessThan(y)
}

func SortByVersion(versions []*Version) {
	sort.Sort(ByVersion(versions))
}

func ConstraintBy(versions []*Version, constraint string) ([]*Version, error) {
	constraint = strings.Replace(constraint, "latest", "", 1)
	if strings.TrimSpace(constraint) == "" {
		constraint = "*"
	}
	if constraint == "*" {
		return versions, nil
	}

	var res []*Version
	var filter func(*Version) bool

	// Regex constraint
	if strings.HasSuffix(constraint, "$") {
		ex, err := regexp.Compile(constraint)
		if err != nil {
			return nil, fmt.Errorf("could not create constraint from regex `%s`: %v", constraint, err)
		}
		filter = func(ver *Version) bool {
			return ex.MatchString(ver.Version)
		}
	} else {
		cond, err := semver.NewConstraint(constraint)
		if err != nil {
			return nil, fmt.Errorf("could not create constraint from string `%s`:\n%v\nSee https://github.com/Masterminds/semver?tab=readme-ov-file#basic-comparisons", constraint, err)
		}
		filter = func(ver *Version) bool {
			v, err := semver.NewVersion(ver.Version)
			if err != nil {
				return false
			}
			return cond.Check(v)
		}
	}
	for _, ver := range versions {
		if filter(ver) {
			res = append(res, ver)
		}
	}
	return res, nil
}
