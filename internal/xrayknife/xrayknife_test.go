package xrayknife

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// FakeRunner is a test double; lives in _test.go so it doesn't ship in the binary.
type FakeRunner struct {
	IsAvailable    bool
	AvailableErr   error
	InstallErr     error
	InstallFlipsAv bool // if true, Install sets IsAvailable=true
	SubsAddErr     error
	SubsFetchErr   error
	HTTPTestErr    error
	SubCountValue  int
	DBPathValue    string
	Calls          []string
}

func (f *FakeRunner) Available(ctx context.Context) (bool, error) {
	f.Calls = append(f.Calls, "Available")
	return f.IsAvailable, f.AvailableErr
}

func (f *FakeRunner) Install(ctx context.Context) error {
	f.Calls = append(f.Calls, "Install")
	if f.InstallErr != nil {
		return f.InstallErr
	}
	if f.InstallFlipsAv {
		f.IsAvailable = true
	}
	return nil
}

func (f *FakeRunner) SubsAdd(ctx context.Context, url, remark string) error {
	f.Calls = append(f.Calls, "SubsAdd:"+remark)
	return f.SubsAddErr
}

func (f *FakeRunner) SubsFetch(ctx context.Context) error {
	f.Calls = append(f.Calls, "SubsFetch")
	return f.SubsFetchErr
}

func (f *FakeRunner) HTTPTest(ctx context.Context, opts HTTPOpts) error {
	tag := "HTTPTest"
	if opts.Speedtest {
		tag = "HTTPTest:speed"
	} else {
		tag = "HTTPTest:ping"
	}
	f.Calls = append(f.Calls, tag)
	return f.HTTPTestErr
}

func (f *FakeRunner) SubCount(ctx context.Context) (int, error) {
	f.Calls = append(f.Calls, "SubCount")
	return f.SubCountValue, nil
}

func (f *FakeRunner) DBPath() (string, error) {
	if f.DBPathValue == "" {
		return "/tmp/fake.db", nil
	}
	return f.DBPathValue, nil
}

func TestEnsureInstalled_AlreadyAvailable(t *testing.T) {
	f := &FakeRunner{IsAvailable: true}
	if err := EnsureInstalled(context.Background(), f); err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	for _, c := range f.Calls {
		if c == "Install" {
			t.Errorf("Install should not be called when already available; calls: %v", f.Calls)
		}
	}
}

func TestEnsureInstalled_InstallsIfMissing(t *testing.T) {
	f := &FakeRunner{IsAvailable: false, InstallFlipsAv: true}
	if err := EnsureInstalled(context.Background(), f); err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	hasInstall := false
	for _, c := range f.Calls {
		if c == "Install" {
			hasInstall = true
		}
	}
	if !hasInstall {
		t.Errorf("Install should have been called; calls: %v", f.Calls)
	}
	if !f.IsAvailable {
		t.Error("IsAvailable should be true after install")
	}
}

func TestEnsureInstalled_FailsIfStillMissing(t *testing.T) {
	// InstallFlipsAv=false simulates a successful install where the binary
	// still doesn't appear on PATH (e.g., $GOPATH/bin missing from PATH).
	f := &FakeRunner{IsAvailable: false, InstallFlipsAv: false}
	err := EnsureInstalled(context.Background(), f)
	if err == nil {
		t.Fatal("expected error when binary still missing after install")
	}
	if !strings.Contains(err.Error(), "PATH") {
		t.Errorf("error should mention PATH; got: %v", err)
	}
}

func TestEnsureInstalled_AvailableErr(t *testing.T) {
	f := &FakeRunner{AvailableErr: errors.New("subprocess crashed")}
	if err := EnsureInstalled(context.Background(), f); err == nil {
		t.Fatal("expected propagated AvailableErr")
	}
}

// TestRealRunner_DBPath sanity-checks the path layout. We don't require it to
// exist, only that it ends in xray-knife.db inside an .xray-knife directory.
func TestRealRunner_DBPath(t *testing.T) {
	r := NewRealRunner()
	p, err := r.DBPath()
	if err != nil {
		t.Fatalf("DBPath: %v", err)
	}
	if !strings.HasSuffix(p, "xray-knife.db") {
		t.Errorf("DBPath should end in xray-knife.db: %q", p)
	}
	if !strings.Contains(p, ".xray-knife") {
		t.Errorf("DBPath should contain .xray-knife dir: %q", p)
	}
}
