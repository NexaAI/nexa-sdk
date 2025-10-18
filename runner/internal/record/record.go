// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

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
		args = []string{"-t", "waveaudio", "-c", "1", "-r", "16000", "-d", outputFile}
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

func (r *Recorder) Run() error {
	return r.cmd.Run()
}

func (r *Recorder) Stop() error {
	if r.cmd.Process == nil {
		return fmt.Errorf("recording not started yet")
	}

	return r.cmd.Process.Kill()
	// _ = r.cmd.Wait()
}

func (r *Recorder) GetOutputFile() string {
	return r.outputFile
}
