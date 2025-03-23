package nix

import (
	"encoding/json"
	"os"
	"os/exec"
	"slices"
)

type JsonMap = map[string]interface{}

var (
	flakes_enabled []string
)

func init() {
	flakes_enabled = []string{
		"--extra-experimental-features",
		"flakes nix-command",
	}
}

func Run(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	cmd.Stderr = nil // capture stdout on err
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func NixRun(args ...string) (string, error) {
	return Run("nix", slices.Concat(flakes_enabled, args)...)
}

func JsonToNix(json string) (string, error) {
	tmpFile, err := os.CreateTemp("", "json-to-nix-*.json")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(json); err != nil {
		return "", err
	}

	// flush
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	return Run(
		"nix-instantiate",
		"--eval",
		"--expr",
		"{f}: builtins.fromJSON (builtins.readFile f)",
		"--arg", "f", tmpFile.Name(),
	)
}

func Nixfmt(args ...string) error {
	_, err := NixRun(
		slices.Concat(
			[]string{"run", "nixpkgs#nixfmt-rfc-style", "--"},
			args,
		)...,
	)
	return err
}

func NixfmtCode(code string) (string, error) {
	tmpFile, err := os.CreateTemp("", "code-*.nix")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		return "", err
	}

	// flush
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	Nixfmt(tmpFile.Name())

	res, err := os.ReadFile(tmpFile.Name())

	return string(res), err
}

func NvJSON(flakePath string) (string, error) {
	return NixRun("eval", "--json", (flakePath + "#lib.nix-versions"))
}

type PackageVersion struct {
	PackageName string
	Version     string
}

func InstallablePackageVersion(installable string) (*PackageVersion, error) {
	out, err := NixRun("eval", "--json", installable, "--apply", "p: { version = p.version; name = builtins.replaceStrings [(\"-\" + p.version)] [\"\"] (p.pname or p.name); }")
	if err != nil {
		return nil, err
	}
	type JsonMap = map[string]interface{}
	obj := JsonMap{}
	err = json.Unmarshal([]byte(out), &obj)
	if err != nil {
		return nil, err
	}
	pv := PackageVersion{
		PackageName: obj["name"].(string),
		Version:     obj["version"].(string),
	}
	return &pv, nil
}
