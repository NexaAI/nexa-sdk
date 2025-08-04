package record

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Recorder struct {
	cmd        *exec.Cmd
	outputFile string
}

func NewRecorder(outputFile string) (*Recorder, error) {
	if _, err := exec.LookPath("sox"); err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	var args []string

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		args = []string{"-d", "-t", "wav", outputFile}
		cmd = exec.Command("sox", args...)

	case "windows":
		args = []string{"-t", "waveaudio", "-d", outputFile}
		cmd = exec.Command("sox", args...)

	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return &Recorder{
		cmd:        cmd,
		outputFile: outputFile,
	}, nil
}

func (r *Recorder) Start() error {
	return r.cmd.Start()
}

func (r *Recorder) Stop() error {
	if r.cmd.Process == nil {
		return fmt.Errorf("recording not start yet")
	}

	return r.cmd.Process.Kill()
	// _ = r.cmd.Wait()
}

func (r *Recorder) GetOutputFile() string {
	return r.outputFile
}
