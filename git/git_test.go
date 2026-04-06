package git

import "testing"

func TestParseCommits(t *testing.T) {
	input := "abc1234\x1fabc\x1ffix login bug\x1fJane Doe\x1f2024-01-15\ndef5678\x1fdef\x1fadd auth middleware\x1fJohn Smith\x1f2024-01-14"
	got := parseCommits(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(got))
	}
	if got[0].Hash != "abc1234" {
		t.Errorf("expected hash abc1234, got %q", got[0].Hash)
	}
	if got[0].ShortHash != "abc" {
		t.Errorf("expected short hash abc, got %q", got[0].ShortHash)
	}
	if got[0].Subject != "fix login bug" {
		t.Errorf("expected subject 'fix login bug', got %q", got[0].Subject)
	}
	if got[0].Author != "Jane Doe" {
		t.Errorf("expected author 'Jane Doe', got %q", got[0].Author)
	}
	if got[0].Date != "2024-01-15" {
		t.Errorf("expected date '2024-01-15', got %q", got[0].Date)
	}
	if got[1].Hash != "def5678" {
		t.Errorf("expected second hash def5678, got %q", got[1].Hash)
	}
}

func TestParseCommitsEmpty(t *testing.T) {
	got := parseCommits("")
	if len(got) != 0 {
		t.Fatalf("expected 0 commits from empty input, got %d", len(got))
	}
}

func TestParseFiles(t *testing.T) {
	input := "M\tsrc/auth.go\nA\tsrc/token.go\nD\tsrc/old.go"
	got := parseFiles(input)
	if len(got) != 3 {
		t.Fatalf("expected 3 files, got %d", len(got))
	}
	if got[0].Status != "M" || got[0].Path != "src/auth.go" {
		t.Errorf("unexpected first file: %+v", got[0])
	}
	if got[1].Status != "A" || got[1].Path != "src/token.go" {
		t.Errorf("unexpected second file: %+v", got[1])
	}
	if got[2].Status != "D" || got[2].Path != "src/old.go" {
		t.Errorf("unexpected third file: %+v", got[2])
	}
}

func TestParseFilesEmpty(t *testing.T) {
	got := parseFiles("")
	if len(got) != 0 {
		t.Fatalf("expected 0 files from empty input, got %d", len(got))
	}
}

func TestParseFilesSkipsBlanks(t *testing.T) {
	input := "M\tsrc/auth.go\n\nA\tsrc/token.go\n"
	got := parseFiles(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 files (blank lines skipped), got %d", len(got))
	}
}
