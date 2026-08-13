package harness_helper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

type fakeStore struct {
	found   *input_itf.HarnessEntity
	findErr error
	saved   *input_itf.HarnessEntity
	saveErr error
}

func (s *fakeStore) Save(info *input_itf.HarnessEntity) error {
	s.saved = info

	return s.saveErr
}

func (s *fakeStore) Find(string) (*input_itf.HarnessEntity, error) {
	return s.found, s.findErr
}

type fakeDownloader struct {
	body []byte
	err  error
}

func (d *fakeDownloader) Download(_, path string, p *input_itf.DownloadParams) error {
	if d.err != nil {
		return d.err
	}

	if p != nil && p.OnProgress != nil {
		p.OnProgress(int64(len(d.body)), int64(len(d.body)))
	}

	return os.WriteFile(path, d.body, 0o644)
}

func (d *fakeDownloader) Reachable(string) error                        { return nil }
func (d *fakeDownloader) GetString(string) (string, error)              { return "", nil }
func (d *fakeDownloader) GetJSON(string, any) error                     { return nil }
func (d *fakeDownloader) PostForm(string, map[string]string, any) error { return nil }
func (d *fakeDownloader) PostJSON(string, any, any) error               { return nil }

func (d *fakeDownloader) Stream(*input_itf.HttpRequest) (*input_itf.HttpResponse, error) {
	return nil, nil
}

func stages(got []input_itf.InstallProgress) []enums.InstallationStage {
	seen := make([]enums.InstallationStage, 0, len(got))

	for _, p := range got {
		if len(seen) == 0 || seen[len(seen)-1] != p.Stage {
			seen = append(seen, p.Stage)
		}
	}

	return seen
}

func specFor(t *testing.T, store *fakeStore, dl *fakeDownloader) (*InstallSpec, string) {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "bin", "tool")

	return &InstallSpec{
		Name:    "tool",
		Label:   "tool",
		BinPath: binPath,
		Store:   store,
		HttpCli: dl,
		Resolve: func() (*Release, error) {
			return &Release{Version: "1.2.3", Platform: "darwin-arm64", URL: "https://example/tool"}, nil
		},
	}, binPath
}

func TestInstallShortCircuitsWhenPresentAndRecorded(t *testing.T) {
	store := &fakeStore{found: &input_itf.HarnessEntity{Name: "tool"}}
	dl := &fakeDownloader{}

	spec, binPath := specFor(t, store, dl)
	spec.Resolve = func() (*Release, error) {
		t.Fatal("resolve must not run for an installed harness")

		return nil, nil
	}

	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(binPath, []byte("already here"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := []input_itf.InstallProgress{}

	if err := Install(spec, func(p input_itf.InstallProgress) { got = append(got, p) }); err != nil {
		t.Fatalf("install: %v", err)
	}

	if want := []enums.InstallationStage{enums.InstallStageDone}; len(stages(got)) != 1 || stages(got)[0] != want[0] {
		t.Fatalf("stages = %v, want %v", stages(got), want)
	}
}

func TestInstallProceedsWhenBinaryPresentButUnrecorded(t *testing.T) {
	store := &fakeStore{}
	dl := &fakeDownloader{body: []byte("fresh")}

	spec, binPath := specFor(t, store, dl)

	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(binPath, []byte("stale"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Install(spec, nil); err != nil {
		t.Fatalf("install: %v", err)
	}

	body, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "fresh" {
		t.Fatalf("binary = %q, want the downloaded one", string(body))
	}
}

func TestInstallSurfacesResolveError(t *testing.T) {
	store := &fakeStore{}
	dl := &fakeDownloader{}

	spec, _ := specFor(t, store, dl)
	spec.Resolve = func() (*Release, error) {
		return nil, custom_error.Critical("no build for this platform")
	}

	err := Install(spec, nil)
	if err == nil || !strings.Contains(err.Error(), "no build") {
		t.Fatalf("err = %v, want the resolve error", err)
	}

	if store.saved != nil {
		t.Fatal("nothing may be recorded when resolve fails")
	}
}

func TestInstallSurfacesDownloadError(t *testing.T) {
	store := &fakeStore{}
	dl := &fakeDownloader{err: custom_error.Critical("connection reset")}

	spec, _ := specFor(t, store, dl)

	err := Install(spec, nil)
	if err == nil || !strings.Contains(err.Error(), "connection reset") {
		t.Fatalf("err = %v, want the download error", err)
	}

	if store.saved != nil {
		t.Fatal("nothing may be recorded when the download fails")
	}
}

func TestInstallWithoutPlaceMovesTheDownloadIntoPlace(t *testing.T) {
	store := &fakeStore{}
	dl := &fakeDownloader{body: []byte("binary")}

	spec, binPath := specFor(t, store, dl)

	got := []input_itf.InstallProgress{}

	if err := Install(spec, func(p input_itf.InstallProgress) { got = append(got, p) }); err != nil {
		t.Fatalf("install: %v", err)
	}

	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("stat binary: %v", err)
	}

	if info.Mode().Perm() != 0o755 {
		t.Fatalf("mode = %v, want 0755", info.Mode().Perm())
	}

	if _, err := os.Stat(binPath + ".download"); err == nil {
		t.Fatal("the temp download must not survive")
	}

	want := []enums.InstallationStage{
		enums.InstallStageResolve,
		enums.InstallStageDownload,
		enums.InstallStageDone,
	}

	if diff := stageDiff(stages(got), want); diff != "" {
		t.Fatal(diff)
	}
}

func TestInstallWithPlaceReportsExtractAndUnpacks(t *testing.T) {
	store := &fakeStore{}
	dl := &fakeDownloader{body: []byte("archive")}

	spec, binPath := specFor(t, store, dl)

	var handed string

	spec.Place = func(downloaded, dest string) error {
		handed = downloaded

		return os.WriteFile(dest, []byte("unpacked"), 0o644)
	}

	got := []input_itf.InstallProgress{}

	if err := Install(spec, func(p input_itf.InstallProgress) { got = append(got, p) }); err != nil {
		t.Fatalf("install: %v", err)
	}

	if handed != binPath+".download" {
		t.Fatalf("place got %q, want the downloaded file", handed)
	}

	body, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "unpacked" {
		t.Fatalf("binary = %q, want the unpacked one", string(body))
	}

	want := []enums.InstallationStage{
		enums.InstallStageResolve,
		enums.InstallStageDownload,
		enums.InstallStageExtract,
		enums.InstallStageDone,
	}

	if diff := stageDiff(stages(got), want); diff != "" {
		t.Fatal(diff)
	}
}

func TestInstallRecordsTheRelease(t *testing.T) {
	store := &fakeStore{}
	dl := &fakeDownloader{body: []byte("binary")}

	spec, binPath := specFor(t, store, dl)

	if err := Install(spec, nil); err != nil {
		t.Fatalf("install: %v", err)
	}

	if store.saved == nil {
		t.Fatal("want the install recorded")
	}

	if store.saved.Version != "1.2.3" {
		t.Fatalf("version = %q, want 1.2.3", store.saved.Version)
	}

	if store.saved.Platform != enums.OS("darwin-arm64") {
		t.Fatalf("platform = %q", store.saved.Platform)
	}

	if store.saved.Path != binPath {
		t.Fatalf("path = %q, want %q", store.saved.Path, binPath)
	}
}

func TestStatusReportsInstalledOnlyWhenTheRecordedPathExists(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "tool")

	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	store := &fakeStore{found: &input_itf.HarnessEntity{Path: binPath, Version: "9.9"}}

	status, err := Status("tool", "Tool", filepath.Join(dir, "missing-cred"), store, 3)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	if !status.Installed || status.Version != "9.9" {
		t.Fatalf("status = %+v, want installed 9.9", status)
	}

	if status.LoggedIn {
		t.Fatal("want logged out with no credential file")
	}

	if status.InstanceCount != 3 {
		t.Fatalf("instances = %d, want 3", status.InstanceCount)
	}

	store.found = &input_itf.HarnessEntity{Path: filepath.Join(dir, "gone"), Version: "9.9"}

	status, err = Status("tool", "Tool", binPath, store, 0)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	if status.Installed {
		t.Fatal("want not installed when the recorded path is gone")
	}

	if !status.LoggedIn {
		t.Fatal("want logged in when the credential file exists")
	}
}

func stageDiff(got, want []enums.InstallationStage) string {
	if len(got) != len(want) {
		return "stages = " + join(got) + ", want " + join(want)
	}

	for i := range got {
		if got[i] != want[i] {
			return "stages = " + join(got) + ", want " + join(want)
		}
	}

	return ""
}

func join(stages []enums.InstallationStage) string {
	parts := make([]string, 0, len(stages))

	for _, s := range stages {
		parts = append(parts, s.String())
	}

	return "[" + strings.Join(parts, " ") + "]"
}
