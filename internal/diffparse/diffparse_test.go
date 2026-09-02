// internal/diffparse/diffparse_test.go
package diffparse

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	diff := "diff --git a/x.go b/x.go\n" +
		"index abc123..def456 100644\n" +
		"--- a/x.go\n" +
		"+++ b/x.go\n" +
		"@@ -1,3 +1,4 @@\n" +
		" package x\n" +
		"-func Old() {}\n" +
		"+func New() {}\n" +
		"+func Extra() {}\n" +
		"\\ No newline at end of file"

	got := Parse(diff)
	want := []Line{
		{Type: LineContext, Content: "diff --git a/x.go b/x.go"},
		{Type: LineContext, Content: "index abc123..def456 100644"},
		{Type: LineHeader, Content: "--- a/x.go"},
		{Type: LineHeader, Content: "+++ b/x.go"},
		{Type: LineHunk, Content: "@@ -1,3 +1,4 @@"},
		{Type: LineContext, Content: " package x"},
		{Type: LineDel, Content: "func Old() {}"},
		{Type: LineAdd, Content: "func New() {}"},
		{Type: LineAdd, Content: "func Extra() {}"},
		{Type: LineContext, Content: "\\ No newline at end of file"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %#v, want %#v", got, want)
	}
}

func TestParseEmpty(t *testing.T) {
	if got := Parse(""); got != nil {
		t.Fatalf("Parse(\"\") = %#v, want nil", got)
	}
}

func TestParseEmptyContextLine(t *testing.T) {
	// A blank line inside a diff hunk is represented as a single space.
	got := Parse("@@ -1,2 +1,2 @@\n \n+added\n")
	want := []Line{
		{Type: LineHunk, Content: "@@ -1,2 +1,2 @@"},
		{Type: LineContext, Content: " "},
		{Type: LineAdd, Content: "added"},
		{Type: LineContext, Content: ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %#v, want %#v", got, want)
	}
}
