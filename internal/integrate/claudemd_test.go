package integrate

import (
	"bytes"
	"testing"
)

func TestMergeClaudeMD_CreateFromEmpty(t *testing.T) {
	out := MergeClaudeMD(nil)
	if !bytes.Equal(out, []byte(ClaudeBlock())) {
		t.Fatalf("from empty should equal the block exactly, got:\n%s", out)
	}
}

func TestMergeClaudeMD_AppendsAndPreserves(t *testing.T) {
	user := []byte("# My project\n\nSome rules here.\n")
	out := MergeClaudeMD(user)
	if !bytes.HasPrefix(out, user) {
		t.Fatal("user content must be preserved as a prefix")
	}
	if !bytes.Contains(out, []byte(claudeStart)) || !bytes.Contains(out, []byte("Some rules here.")) {
		t.Fatal("expected both the block and the original text")
	}
	// blank-line separator between user content and the block
	if !bytes.Contains(out, []byte("\n\n"+claudeStart)) {
		t.Fatal("expected a blank line before the inserted block")
	}
}

func TestMergeClaudeMD_Idempotent(t *testing.T) {
	user := []byte("# My project\n\nSome rules here.\n")
	once := MergeClaudeMD(user)
	twice := MergeClaudeMD(once)
	if !bytes.Equal(once, twice) {
		t.Fatalf("merge must be idempotent\nonce:\n%s\ntwice:\n%s", once, twice)
	}
	if n := bytes.Count(twice, []byte(claudeStart)); n != 1 {
		t.Fatalf("expected exactly one start marker, got %d", n)
	}
}

func TestMergeClaudeMD_RefreshesBetweenMarkers(t *testing.T) {
	stale := []byte("# Top\n\n" + claudeStart + "\nOLD STALE TEXT\n" + claudeEnd + "\n\n# Bottom\n")
	out := MergeClaudeMD(stale)
	if bytes.Contains(out, []byte("OLD STALE TEXT")) {
		t.Fatal("stale block content must be replaced")
	}
	if !bytes.Contains(out, []byte("# Top")) || !bytes.Contains(out, []byte("# Bottom")) {
		t.Fatal("content surrounding the markers must be preserved")
	}
	if n := bytes.Count(out, []byte(claudeStart)); n != 1 {
		t.Fatalf("expected one start marker, got %d", n)
	}
	if !bytes.Equal(out, MergeClaudeMD(out)) {
		t.Fatal("refreshed file must be idempotent")
	}
}

func TestMergeClaudeMD_NoTrailingNewline(t *testing.T) {
	in := []byte("line without newline")
	out := MergeClaudeMD(in)
	if !bytes.HasPrefix(out, in) {
		t.Fatal("must preserve content lacking a trailing newline")
	}
	if !bytes.Contains(out, []byte("\n\n"+claudeStart)) {
		t.Fatal("must insert a blank-line separator even without a trailing newline")
	}
}
