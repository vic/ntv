package find

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

	nixsearch "github.com/peterldowns/nix-search-cli/pkg/nixsearch"
	"github.com/vic/nix-versions/packages/lazamar"
	"github.com/vic/nix-versions/packages/nixhub"
	lib "github.com/vic/nix-versions/packages/versions"
)

type Opts struct {
	One            bool
	Exact          bool
	Constraint     string
	Limit          int
	Sort           bool
	Reverse        bool
	Lazamar        bool
	Channel        string
	ignoreFetchErr bool
}

var (
	RE_NIX_TOOL              *regexp.Regexp
	RE_NIX_INSTALLABLE       *regexp.Regexp
	RE_NIX_INSTALLABLE_SHORT *regexp.Regexp
	TOOL_FILE_READERS        map[string](func(string) (*string, error))
	TOOLS_FILE_READERS       map[string](func(string) ([]string, error))
)

func init() {
	RE_NIX_TOOL = regexp.MustCompile(`^([^ #]+[^ ]+)[ ]*(# ([^ #]+)@([^ #]+))?`)
	RE_NIX_INSTALLABLE = regexp.MustCompile(`^([^#]+/)([^#]+)#([^# ]+)(#([^ #]+)#([^ #]+))?`)
	RE_NIX_INSTALLABLE_SHORT = regexp.MustCompile(`^([^#/]+)#([^# ]+)(#([^ #]+)#([^ #]+))?`)

	TOOL_FILE_READERS = map[string](func(string) (*string, error)){
		".node-version": readConstraint,
		".java-version": readConstraint,
		".ruby-version": readConstraint,
	}

	TOOLS_FILE_READERS = map[string](func(string) ([]string, error)){
		".nix-tools":     readNixTools,
		".tool-versions": readAsdfToolVersions,
	}
}

func FindPackagesWithQuery(ctx Opts, search string) ([]string, error) {
	var limit int
	if ctx.Limit < 0 {
		limit = max(ctx.Limit*-1, 10)
	} else {
		limit = max(ctx.Limit, 10)
	}
	query := nixsearch.Query{
		MaxResults: limit,
		Channel:    "unstable",
		Search:     &nixsearch.MatchSearch{Search: search},
	}
	client, err := nixsearch.NewElasticSearchClient()
	if err != nil {
		return nil, err
	}
	packages, err := client.Search(context.Background(), query)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, pkg := range packages {
		names = append(names, pkg.AttrName)
	}
	if len(names) < 1 {
		return nil, fmt.Errorf("No packages found matching `%s`.", search)
	}
	return names, nil
}

func FindPackagesWithProgram(ctx Opts, program string) ([]string, error) {
	var limit int
	if ctx.Limit < 0 {
		limit = max(ctx.Limit*-1, 10)
	} else {
		limit = max(ctx.Limit, 10)
	}
	query := nixsearch.Query{
		MaxResults: limit,
		Channel:    "unstable",
		Program:    &nixsearch.MatchProgram{Program: program},
	}
	client, err := nixsearch.NewElasticSearchClient()
	if err != nil {
		return nil, err
	}
	packages, err := client.Search(context.Background(), query)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, pkg := range packages {
		if ctx.Exact {
			if slices.Contains(pkg.Programs, program) {
				names = append(names, pkg.AttrName)
			}
		} else {
			names = append(names, pkg.AttrName)
		}
	}
	if len(names) < 1 {
		return nil, fmt.Errorf("No packages found providing program `bin/%s`.\nTry using `--exact=false` option to match on any part of the program name.", program)
	}
	return names, nil
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func readConstraint(path string) (*string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	constraint := strings.TrimSpace(string(bytes))
	return &constraint, nil
}

func readAsdfToolVersions(file string) ([]string, error) {
	return nil, nil
}

func readNixTools(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var names []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		match := RE_NIX_TOOL.FindStringSubmatch(line)
		if len(match) > 1 {
			name := match[1]
			if isInstallable(name) {
				if len(match) > 4 && match[4] != "" {
					name = name + "#" + match[3] + "#" + match[4]
				}
				names = append(names, name)
			} else {
				read, err := readPackagesFromFile(name)
				if err != nil {
					return nil, err
				}
				names = append(names, read...)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

func readPackagesFromFile(arg string) ([]string, error) {
	if strings.Contains(arg, "@") && isFile(arg[strings.Index(arg, "@")+1:]) {
		name := arg[:strings.Index(arg, "@")]
		file := arg[strings.Index(arg, "@")+1:]

		tools_reader, is_tools_file := TOOLS_FILE_READERS[path.Base(file)]
		tool_reader, is_tool_file := TOOL_FILE_READERS[path.Base(file)]

		if name == "" {
			if is_tools_file {
				return tools_reader(file)
			}
			return readNixTools(file)
		}

		if is_tool_file {
			constraint, err := tool_reader(file)
			if err != nil {
				return nil, err
			}
			return []string{name + "@" + *constraint}, nil
		}

		tools_reader, is_tools_file = TOOLS_FILE_READERS[name]
		if is_tools_file {
			return tools_reader(file)
		}

		constraint, err := readConstraint(file)
		if err != nil {
			return nil, err
		}
		return []string{name + "@" + *constraint}, nil
	}
	return []string{arg}, nil
}

func isInstallable(str string) bool {
	return strings.ContainsAny(str, ":/#") && !strings.HasPrefix(str, "bin/")
}

func fromInstallableStr(str string) lib.Version {
	var (
		flake    string
		revision string
		attr     string
		version  string
		name     string
	)
	match := RE_NIX_INSTALLABLE.FindStringSubmatch(str)
	if len(match) > 0 {
		if len(match) > 6 {
			version = match[6]
		}
		if len(match) > 5 {
			name = match[5]
		}
		if len(match) > 3 {
			attr = match[3]
		}
		if len(match) > 2 {
			revision = match[2]
		}
		flake = strings.TrimRight(match[1], "/")
	} else {
		match := RE_NIX_INSTALLABLE_SHORT.FindStringSubmatch(str)
		if len(match) > 5 {
			version = match[5]
		}
		if len(match) > 4 {
			name = match[4]
		}
		if len(match) > 2 {
			attr = match[2]
		}
		flake = match[1]
		revision = "HEAD"
	}
	result := lib.Version{
		Name:      name,
		Version:   version,
		Flake:     flake,
		Revision:  revision,
		Attribute: attr,
	}
	return result
}

func FindVersions(ctx Opts, name string) ([]lib.Version, error) {
	if isInstallable(name) {
		return []lib.Version{fromInstallableStr(name)}, nil
	}

	var (
		err              error
		versions         []lib.Version
		constraint       = ctx.Constraint
		limit            = ctx.Limit
		sort             = ctx.Sort
		reverse          = ctx.Reverse
		pkgAttr          = name
		constraintInName = strings.Contains(name, "@")
	)
	if constraintInName {
		pkgAttr = name[:strings.Index(name, "@")]
		constraint = name[strings.Index(name, "@")+1:]
	}
	searchAgain := func(attrs []string, err error) ([]lib.Version, error) {
		if err != nil {
			return nil, err
		}
		var pkgs []string
		if constraintInName {
			for _, attr := range attrs {
				pkgs = append(pkgs, attr+"@"+constraint)
			}
		} else {
			pkgs = attrs
		}
		opts := Opts{
			One:            ctx.One,
			Exact:          ctx.Exact,
			Constraint:     ctx.Constraint,
			Limit:          ctx.Limit,
			Sort:           ctx.Sort,
			Reverse:        ctx.Reverse,
			Lazamar:        ctx.Lazamar,
			Channel:        ctx.Channel,
			ignoreFetchErr: true,
		}
		return FindVersionsAll(opts, pkgs)
	}
	if strings.HasPrefix(pkgAttr, "~") {
		attrs, err := FindPackagesWithQuery(ctx, pkgAttr[1:])
		return searchAgain(attrs, err)
	}
	if strings.HasPrefix(pkgAttr, "bin/") {
		attrs, err := FindPackagesWithProgram(ctx, pkgAttr[4:])
		return searchAgain(attrs, err)
	}
	if strings.Contains(constraint, "latest") {
		constraint = strings.Replace(constraint, "latest", "", 1)
		limit = 1
		sort = true
		reverse = false
	}
	if ctx.Lazamar {
		versions, err = lazamar.Versions(pkgAttr, ctx.Channel)
	} else {
		versions, err = nixhub.Versions(pkgAttr)
	}
	if err != nil {
		if ctx.ignoreFetchErr {
			return versions, nil
		}
		return nil, err
	}
	if ctx.Exact {
		versions = lib.Exact(versions, pkgAttr)
	}
	if constraint != "" {
		versions, err = lib.ConstraintBy(versions, constraint)
		if err != nil {
			return nil, err
		}
	}
	if sort {
		lib.SortByVersion(versions)
	}
	if reverse {
		slices.Reverse(versions)
	}
	if limit != 0 {
		versions = lib.Limit(versions, limit)
	}
	return versions, nil
}

func FindVersionsAll(ctx Opts, names []string) ([]lib.Version, error) {
	var versions []lib.Version

	for _, name := range names {
		vers, err := FindVersions(ctx, name)
		if err != nil {
			return nil, err
		}
		versions = append(versions, vers...)
	}

	if ctx.Sort {
		lib.SortByVersion(versions)
	}
	if ctx.Reverse {
		slices.Reverse(versions)
	}
	if ctx.Limit != 0 {
		versions = lib.Limit(versions, ctx.Limit)
	}

	return versions, nil
}

func FindReadingVersionsAll(ctx Opts, args []string) ([]lib.Version, error) {
	var names []string
	for _, arg := range args {
		namex, err := readPackagesFromFile(arg)
		if err != nil {
			return nil, err
		}
		names = append(names, namex...)
	}
	return FindVersionsAll(ctx, names)
}
