package list

import (
	"testing"
)

func assert(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Errorf("assertion failed: %s", msg)
	}
}
func assertNoErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSpecFromLine_empty(t *testing.T) {
	spec, err := specFromLine("")
	assertNoErr(t, err)
	assert(t, spec == "", "should be empty")
}

func TestSpecFromLine_ignore_comment(t *testing.T) {
	spec, err := specFromLine("# comment")
	assertNoErr(t, err)
	assert(t, spec == "", "should be empty")
}

func TestSpecFromLine_simple(t *testing.T) {
	spec, err := specFromLine("foo # comment")
	assertNoErr(t, err)
	assert(t, spec == "foo", "simple")
}

func TestSpecFromLine_allow_spaces_in_spec(t *testing.T) {
	spec, err := specFromLine("  emacs@ >27 || <29 # comment")
	assertNoErr(t, err)
	assert(t, spec == "emacs@ >27 || <29", "emacs")
}

func TestSpecFromLine_asdf_compatible(t *testing.T) {
	spec, err := specFromLine("emacs 27 || 29")
	assertNoErr(t, err)
	assert(t, spec == "emacs@27 || 29", "should replace first space by @")
}

func TestSpecFromLine_asdf_compatible_multi_spaces(t *testing.T) {
	spec, err := specFromLine("emacs  	 	  27 || 29")
	assertNoErr(t, err)
	assert(t, spec == "emacs@27 || 29", "should replace first space by @")
}

func TestSpecFromLine_installable(t *testing.T) {
	spec, err := specFromLine("foo#bar^out,lib 25 #comment")
	assertNoErr(t, err)
	assert(t, spec == "foo#bar^out,lib@25", "should replace first space by @")
}
