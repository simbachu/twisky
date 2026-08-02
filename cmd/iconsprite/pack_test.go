package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackRawSVG(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "raw.svg"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	out, err := pack(raw)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	got := string(out)

	if !strings.Contains(got, `xmlns="http://www.w3.org/2000/svg"`) {
		t.Fatalf("missing svg xmlns:\n%s", got)
	}
	if !strings.Contains(got, `style="display:none"`) {
		t.Fatalf("missing display:none on root:\n%s", got)
	}
	if !strings.Contains(got, `<symbol id="icon-alpha" viewBox="0 0 64 64">`) {
		t.Fatalf("missing icon-alpha symbol:\n%s", got)
	}
	if !strings.Contains(got, `<symbol id="icon-beta" viewBox="0 0 64 64">`) {
		t.Fatalf("missing icon-beta symbol (page label, geometry match):\n%s", got)
	}
	if strings.Count(got, `<symbol`) != 2 {
		t.Fatalf("expected exactly 2 symbols (skip unlabeled page):\n%s", got)
	}
	if strings.Contains(got, "inkscape:") || strings.Contains(got, "sodipodi:") {
		t.Fatalf("output still contains inkscape/sodipodi chrome:\n%s", got)
	}
	if !strings.Contains(got, `d="M 8,8 L 56,56"`) {
		t.Fatalf("missing alpha path data:\n%s", got)
	}
	if !strings.Contains(got, `translate(-68`) {
		t.Fatalf("expected page-local translate for beta (page x=68):\n%s", got)
	}
}

func TestPackEmptyLabeledPage(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "empty_labeled.svg"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	_, err = pack(raw)
	if err == nil {
		t.Fatal("expected error for labeled page with no drawable content")
	}
	if !strings.Contains(err.Error(), "icon-lonely") {
		t.Fatalf("error should mention page label, got: %v", err)
	}
}

func TestPackDuplicateLabels(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "duplicate_labels.svg"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	_, err = pack(raw)
	if err == nil {
		t.Fatal("expected error for duplicate page labels")
	}
	if !strings.Contains(err.Error(), "icon-dup") {
		t.Fatalf("error should mention duplicate label, got: %v", err)
	}
}
