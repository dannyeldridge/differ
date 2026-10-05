package gui

import "testing"

func TestRepoFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"separate value", []string{"--repo", "/a/b"}, "/a/b"},
		{"equals form", []string{"--repo=/a/b"}, "/a/b"},
		{"single dash", []string{"-repo", "/a/b"}, "/a/b"},
		{"among other flags", []string{"--foreground", "--repo", "/a/b"}, "/a/b"},
		{"missing", []string{"--foreground"}, ""},
		{"flag without value", []string{"--repo"}, ""},
		{"empty", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repoFromArgs(tt.args); got != tt.want {
				t.Fatalf("repoFromArgs(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}
