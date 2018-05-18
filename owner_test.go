package dreck

import (
	"io/ioutil"
	golog "log"
	"testing"

	"github.com/google/go-github/github"
)

var _ = func() bool {
	golog.SetOutput(ioutil.Discard)
	return true
}()

func TestOwnersSingle(t *testing.T) {
	d := New()

	files := []*github.CommitFile{
		&github.CommitFile{Filename: String("/home/example/test.txt")},
	}
	victim, _ := d.findReviewers(files, "ab", func(path string) ([]byte, error) {
		return []byte(`reviewers:
- ab
- ac
`), nil
	})

	if expect := "ac"; victim != expect {
		t.Errorf("expected %s, got %s", expect, victim)
	}
}

func TestOwnersMultiple(t *testing.T) {
	d := New()

	files := []*github.CommitFile{
		&github.CommitFile{Filename: String("/home/example/a/test.txt")},
		&github.CommitFile{Filename: String("/home/example/test.txt")},
	}
	victim, _ := d.findReviewers(files, "ac", func(path string) ([]byte, error) {
		switch path {
		case "/home/example/a/OWNERS":
			return []byte(`reviewers:
- ab
- ac
`), nil
		case "/home/example/OWNERS":
			return []byte(`reviewers:
- xb
- xc
`), nil
		}
		return nil, nil
	})

	// ac is the puller
	if expect := "ab"; victim != expect {
		t.Errorf("expected %s, got %s", expect, victim)
	}

	victim, _ = d.findReviewers(files, "ac", func(path string) ([]byte, error) {
		switch path {
		case "/home/example/a/OWNERS":
			return []byte(`reviewers:
- ac
`), nil
		case "/home/example/OWNERS":
			return []byte(`reviewers:
- xb
`), nil
		}
		return nil, nil
	})

	// ac is the puller, but can't be selected, so xb should be it.
	if expect := "xb"; victim != expect {
		t.Errorf("expected %s, got %s", expect, victim)
	}
}

func TestOwnersMostSpecific(t *testing.T) {
	d := New()

	files := []*github.CommitFile{
		&github.CommitFile{Filename: String("/home/plugin/reload/test.txt")},
	}
	victim, _ := d.findReviewers(files, "aa", func(path string) ([]byte, error) {
		switch path {
		case "/home/plugin/reload/OWNERS":
			return []byte(`reviewers:
- aa
`), nil
		case "/home/plugin/OWNERS":
			return []byte(`reviewers:
- bb
`), nil
		}
		return nil, nil
	})

	if expect := "bb"; victim != expect {
		t.Errorf("expected %s, got %s", expect, victim)
	}
}
