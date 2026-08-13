package harness_helper

import (
	"os"
	"path/filepath"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

type Release struct {
	Version  string
	Platform string
	URL      string
	Checksum string
}

type InstallSpec struct {
	Name    string
	Label   string
	BinPath string
	Store   input_itf.HarnessStorage
	HttpCli input_itf.HttpCli

	Resolve func() (*Release, error)

	Place func(downloaded, dest string) error
}

func Install(spec *InstallSpec, onProgress func(input_itf.InstallProgress)) error {
	if onProgress == nil {
		onProgress = func(input_itf.InstallProgress) {}
	}

	installed, err := spec.alreadyInstalled()
	if err != nil {
		return err
	}

	if installed {
		onProgress(input_itf.InstallProgress{Stage: enums.InstallStageDone})

		return nil
	}

	onProgress(input_itf.InstallProgress{Stage: enums.InstallStageResolve})

	release, err := spec.Resolve()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(spec.BinPath), 0o755); err != nil {
		return custom_error.Critical("create %s bin dir: %v", spec.Label, err)
	}

	downloaded := spec.BinPath + ".download"
	defer os.Remove(downloaded)

	onProgress(input_itf.InstallProgress{Stage: enums.InstallStageDownload})

	if err := spec.HttpCli.Download(release.URL, downloaded, &input_itf.DownloadParams{
		Checksum: release.Checksum,
		OnProgress: func(bytes, total int64) {
			onProgress(input_itf.InstallProgress{
				Stage:      enums.InstallStageDownload,
				Downloaded: bytes,
				Total:      total,
			})
		},
	}); err != nil {
		return custom_error.Critical("download %s: %v", spec.Label, err)
	}

	if err := spec.place(downloaded, onProgress); err != nil {
		return err
	}

	if err := spec.Store.Save(&input_itf.HarnessEntity{
		Name:     spec.Name,
		Version:  release.Version,
		Platform: enums.OS(release.Platform),
		Path:     spec.BinPath,
	}); err != nil {
		return custom_error.Critical("save install info: %v", err)
	}

	onProgress(input_itf.InstallProgress{Stage: enums.InstallStageDone})

	return nil
}

func (spec *InstallSpec) alreadyInstalled() (bool, error) {
	if _, err := os.Stat(spec.BinPath); err != nil {
		return false, nil
	}

	info, err := spec.Store.Find(spec.Name)
	if err != nil {
		return false, custom_error.Critical("find harness info: %v", err)
	}

	return info != nil, nil
}

func (spec *InstallSpec) place(downloaded string, onProgress func(input_itf.InstallProgress)) error {
	if spec.Place == nil {
		if err := os.Rename(downloaded, spec.BinPath); err != nil {
			return custom_error.Critical("install %s binary: %v", spec.Label, err)
		}

		return chmodExec(spec.BinPath, spec.Label)
	}

	onProgress(input_itf.InstallProgress{Stage: enums.InstallStageExtract})

	if err := spec.Place(downloaded, spec.BinPath); err != nil {
		return err
	}

	return chmodExec(spec.BinPath, spec.Label)
}

func chmodExec(path, label string) error {
	if err := os.Chmod(path, 0o755); err != nil {
		return custom_error.Critical("make %s binary executable: %v", label, err)
	}

	return nil
}

func Status(name, display, credPath string, store input_itf.HarnessStorage, instances int) (*input_itf.AgentStatus, error) {
	status := &input_itf.AgentStatus{Name: display, InstanceCount: instances}

	info, err := store.Find(name)
	if err != nil {
		return nil, custom_error.Critical("find harness info: %v", err)
	}

	if info != nil && exists(info.Path) {
		status.Installed = true
		status.Version = info.Version
	}

	status.LoggedIn = exists(credPath)

	return status, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
