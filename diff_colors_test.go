package main

import "testing"

func TestColorizeDiffPreservesContent(t *testing.T) {
	diff := "--- a/x.go\n" +
		"+++ b/x.go\n" +
		"@@ -1,2 +1,2 @@\n" +
		" package x\n" +
		"-func Old() {}\n" +
		"+func New() {}\n"

	got := ColorizeDiff(diff)
	// strings.Split(diff, "\n") followed by strings.Join(_, "\n") is a lossless
	// round-trip, trailing newline included, so the unstyled output must equal
	// the input exactly.
	if got != diff {
		t.Fatalf("ColorizeDiff() = %q, want %q", got, diff)
	}
}
