package find

import (
	"os"
	"testing"
)

func Test_isFile_existing(t *testing.T) {
	if !isFile("testdata/emacs-version") {
		t.Error("should be a file")
	}
}

func Test_isFile_nonexistent(t *testing.T) {
	if isFile("testdata/emacs-version-nonexistent") {
		t.Error("should not be a file")
	}
}

func Test_readConstraint_trims(t *testing.T) {
	ptr, err := readConstraint("testdata/emacs-version")
	if err != nil {
		t.Error(err)
		return
	}
	if *ptr != "~25" {
		t.Error("should be ~25", *ptr)
	}
}

func Test_readPackagesFromFile_at_node_version(t *testing.T) {
	arr, err := readPackagesFromFile("node@testdata/.node-version")
	if err != nil {
		t.Error(err)
		return
	}
	if arr[0] != "node@latest" {
		t.Error("should be node@latest", arr[0])
	}
}

func Test_readPackagesFromFile_at_nonfile(t *testing.T) {
	arr, err := readPackagesFromFile("node@<23")
	if err != nil {
		t.Error(err)
		return
	}
	if arr[0] != "node@<23" {
		t.Error("should be node@<23", arr[0])
	}
}

func Test_readPackagesFromFile_nix_tools(t *testing.T) {
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	err := os.Chdir("testdata")
	if err != nil {
		t.Error(err)
		return
	}
	arr, err := readPackagesFromFile("@nix-tools")
	if err != nil {
		t.Error(err)
		return
	}
	if len(arr) != 1 {
		t.Error(len(arr))
	}
	if arr[0] != "emacs-nox@~25" {
		t.Error(arr[0])
	}
}

func Test_readPackagesFromFile_dot_nix_tools(t *testing.T) {
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	err := os.Chdir("testdata")
	if err != nil {
		t.Error(err)
		return
	}
	arr, err := readPackagesFromFile("@.nix-tools")
	if err != nil {
		t.Error(err)
		return
	}
	if len(arr) != 6 {
		t.Error(len(arr))
	}
	if arr[0] != "nixpkgs/0d534853a55b5d02a4ababa1d71921ce8f0aee4c#cargo#cargo#1.85.0" {
		t.Error(arr[0])
	}
	if arr[1] != "nixpkgs/0d534853a55b5d02a4ababa1d71921ce8f0aee4c#rustc" {
		t.Error(arr[1])
	}
	if arr[2] != "nixpkgs/master#hello" {
		t.Error(arr[2])
	}
	if arr[3] != "nixpkgs#btop#btop#1.0" {
		t.Error(arr[3])
	}
	if arr[4] != "bin/emacs@~25" {
		t.Error(arr[4])
	}
	if arr[5] != "emacs-nox@~25" {
		t.Error(arr[5])
	}
}

func Test_isInstallable_head_hello(t *testing.T) {
	if !isInstallable("nixpkgs/HEAD#hello") {
		t.Error("should be installable")
	}
}

func Test_isInstallable_default_app(t *testing.T) {
	if !isInstallable("github:vic/nix-versions") {
		t.Error("should be installable")
	}
}

func Test_fromInstallableStr_preserves_meta_version(t *testing.T) {
	v := fromInstallableStr("nixpkgs/HEAD#some.hello#hello#1.0.0")
	if v.Name != "hello" {
		t.Error("should preserve name", v)
	}
	if v.Version != "1.0.0" {
		t.Error("should preserve version", v)
	}
	if v.Attribute != "some.hello" {
		t.Error("should preserve attribute", v)
	}
	if v.Revision != "HEAD" {
		t.Error("should preserve revision", v)
	}
	if v.Flake != "nixpkgs" {
		t.Error("should preserve flake", v)
	}
}

func Test_fromInstallableStr_with_no_meta_version(t *testing.T) {
	v := fromInstallableStr("github:nixos/nixpkgs/master#hello")
	if v.Version != "" {
		t.Error("should preserve version", v)
	}
	if v.Attribute != "hello" {
		t.Error("should preserve attribute", v)
	}
	if v.Revision != "master" {
		t.Error("should preserve revision", v)
	}
	if v.Flake != "github:nixos/nixpkgs" {
		t.Error("should preserve flake", v)
	}
}

func Test_fromInstallableStr_with_registry(t *testing.T) {
	v := fromInstallableStr("nixpkgs#hello")
	if v.Version != "" {
		t.Error("should preserve version", v)
	}
	if v.Attribute != "hello" {
		t.Error("should preserve attribute", v)
	}
	if v.Revision != "HEAD" {
		t.Error("should preserve revision", v)
	}
	if v.Flake != "nixpkgs" {
		t.Error("should preserve flake", v)
	}
}

func Test_fromInstallableStr_with_registry_and_meta_version(t *testing.T) {
	v := fromInstallableStr("nixpkgs#hello#hey#3.0")
	if v.Name != "hey" {
		t.Error("should preserve name", v)
	}
	if v.Version != "3.0" {
		t.Error("should preserve version", v)
	}
	if v.Attribute != "hello" {
		t.Error("should preserve attribute", v)
	}
	if v.Revision != "HEAD" {
		t.Error("should preserve revision", v)
	}
	if v.Flake != "nixpkgs" {
		t.Error("should preserve flake", v)
	}
}
