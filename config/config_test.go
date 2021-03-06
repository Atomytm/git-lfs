package config

import (
	"os"
	"testing"

	"github.com/git-lfs/git-lfs/git"
	"github.com/stretchr/testify/assert"
)

func TestRemoteDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.unused.remote":     []string{"a"},
			"branch.unused.pushRemote": []string{"b"},
		},
	})
	assert.Equal(t, "origin", cfg.Remote())
	assert.Equal(t, "origin", cfg.PushRemote())
}

func TestRemoteBranchConfig(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":    []string{"a"},
			"branch.other.pushRemote": []string{"b"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "a", cfg.PushRemote())
}

func TestRemotePushDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":    []string{"a"},
			"remote.pushDefault":      []string{"b"},
			"branch.other.pushRemote": []string{"c"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "b", cfg.PushRemote())
}

func TestRemoteBranchPushDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":     []string{"a"},
			"remote.pushDefault":       []string{"b"},
			"branch.master.pushRemote": []string{"c"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "c", cfg.PushRemote())
}

func TestBasicTransfersOnlySetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.basictransfersonly": []string{"true"},
		},
	})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, true, b)
}

func TestBasicTransfersOnlyDefault(t *testing.T) {
	cfg := NewFrom(Values{})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, false, b)
}

func TestBasicTransfersOnlyInvalidValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.basictransfersonly": []string{"wat"},
		},
	})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, false, b)
}

func TestTusTransfersAllowedSetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.tustransfers": []string{"true"},
		},
	})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, true, b)
}

func TestTusTransfersAllowedDefault(t *testing.T) {
	cfg := NewFrom(Values{})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, false, b)
}

func TestTusTransfersAllowedInvalidValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.tustransfers": []string{"wat"},
		},
	})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, false, b)
}

func TestLoadValidExtension(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.extension.foo.clean":    []string{"foo-clean %f"},
			"lfs.extension.foo.smudge":   []string{"foo-smudge %f"},
			"lfs.extension.foo.priority": []string{"2"},
		},
	})

	ext := cfg.Extensions()["foo"]

	assert.Equal(t, "foo", ext.Name)
	assert.Equal(t, "foo-clean %f", ext.Clean)
	assert.Equal(t, "foo-smudge %f", ext.Smudge)
	assert.Equal(t, 2, ext.Priority)
}

func TestLoadInvalidExtension(t *testing.T) {
	cfg := NewFrom(Values{})
	ext := cfg.Extensions()["foo"]

	assert.Equal(t, "", ext.Name)
	assert.Equal(t, "", ext.Clean)
	assert.Equal(t, "", ext.Smudge)
	assert.Equal(t, 0, ext.Priority)
}

func TestFetchIncludeExcludesAreCleaned(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.fetchinclude": []string{"/path/to/clean/"},
			"lfs.fetchexclude": []string{"/other/path/to/clean/"},
		},
	})

	assert.Equal(t, []string{"/path/to/clean"}, cfg.FetchIncludePaths())
	assert.Equal(t, []string{"/other/path/to/clean"}, cfg.FetchExcludePaths())
}

func TestRepositoryPermissions(t *testing.T) {
	perms := 0666 & ^umask()

	values := map[string]int{
		"group":     0660,
		"true":      0660,
		"1":         0660,
		"YES":       0660,
		"all":       0664,
		"world":     0664,
		"everybody": 0664,
		"2":         0664,
		"false":     perms,
		"umask":     perms,
		"0":         perms,
		"NO":        perms,
		"this does not remotely look like a valid value": perms,
		"0664": 0664,
		"0666": 0666,
		"0600": 0600,
		"0660": 0660,
		"0644": 0644,
	}

	for key, val := range values {
		cfg := NewFrom(Values{
			Git: map[string][]string{
				"core.sharedrepository": []string{key},
			},
		})
		assert.Equal(t, os.FileMode(val), cfg.RepositoryPermissions())
	}
}

func TestCurrentUser(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"user.name":  []string{"Pat Doe"},
			"user.email": []string{"pdoe@example.org"},
		},
		Os: map[string][]string{
			"EMAIL": []string{"pdoe@example.com"},
		},
	})

	name, email := cfg.CurrentCommitter()
	assert.Equal(t, name, "Pat Doe")
	assert.Equal(t, email, "pdoe@example.org")

	cfg = NewFrom(Values{
		Git: map[string][]string{
			"user.name": []string{"Pat Doe"},
		},
		Os: map[string][]string{
			"EMAIL": []string{"pdoe@example.com"},
		},
	})

	name, email = cfg.CurrentCommitter()
	assert.Equal(t, name, "Pat Doe")
	assert.Equal(t, email, "pdoe@example.com")

	cfg = NewFrom(Values{
		Git: map[string][]string{
			"user.name":  []string{"Pat Doe"},
			"user.email": []string{"pdoe@example.org"},
		},
		Os: map[string][]string{
			"GIT_COMMITTER_NAME":  []string{"Sam Roe"},
			"GIT_COMMITTER_EMAIL": []string{"sroe@example.net"},
			"EMAIL":               []string{"pdoe@example.com"},
		},
	})

	name, email = cfg.CurrentCommitter()
	assert.Equal(t, name, "Sam Roe")
	assert.Equal(t, email, "sroe@example.net")

	cfg = NewFrom(Values{
		Git: map[string][]string{
			"user.name":  []string{"Pat Doe"},
			"user.email": []string{"pdoe@example.org"},
		},
		Os: map[string][]string{
			"GIT_AUTHOR_NAME":  []string{"Sam Roe"},
			"GIT_AUTHOR_EMAIL": []string{"sroe@example.net"},
			"EMAIL":            []string{"pdoe@example.com"},
		},
	})

	name, email = cfg.CurrentCommitter()
	assert.Equal(t, name, "Pat Doe")
	assert.Equal(t, email, "pdoe@example.org")

	name, email = cfg.CurrentAuthor()
	assert.Equal(t, name, "Sam Roe")
	assert.Equal(t, email, "sroe@example.net")
}
