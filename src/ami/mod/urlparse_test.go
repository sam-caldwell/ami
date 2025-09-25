package mod

import "testing"

func TestParseGitPlusSSH_Valid(t *testing.T) {
    raw := "git+ssh://github.com/org/repo.git#v1.2.3"
    got, err := parseGitPlusSSH(raw)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if got.SSHURL != "ssh://git@github.com/org/repo.git" { t.Fatalf("sshurl: %q", got.SSHURL) }
    if got.Host != "github.com" { t.Fatalf("host: %q", got.Host) }
    if got.Path != "/org/repo.git" { t.Fatalf("path: %q", got.Path) }
    if got.Repo != "repo" { t.Fatalf("repo: %q", got.Repo) }
    if got.Tag != "v1.2.3" { t.Fatalf("tag: %q", got.Tag) }
}

func TestParseGitPlusSSH_AppendsDotGitAndSemverCheck(t *testing.T) {
    raw := "git+ssh://example.com/acme/widgets#v0.9.0"
    got, err := parseGitPlusSSH(raw)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if got.Path != "/acme/widgets.git" { t.Fatalf("path: %q", got.Path) }
    if got.Repo != "widgets" { t.Fatalf("repo: %q", got.Repo) }
    if got.Tag != "v0.9.0" { t.Fatalf("tag: %q", got.Tag) }
}

func TestParseGitPlusSSH_MissingTag_Error(t *testing.T) {
    raw := "git+ssh://example.com/acme/widgets"
    if _, err := parseGitPlusSSH(raw); err == nil {
        t.Fatalf("expected error for missing tag")
    }
}

func TestParseGitPlusSSH_InvalidSemverTag_Error(t *testing.T) {
    raw := "git+ssh://example.com/acme/widgets#dev"
    if _, err := parseGitPlusSSH(raw); err == nil {
        t.Fatalf("expected error for invalid semver tag")
    }
}

