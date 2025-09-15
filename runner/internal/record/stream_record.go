package record

import (
	"fmt"
	"io"
	"os/exec"
)

type StreamRecorder struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func NewStreamRecorder() (*StreamRecorder, error) {
	return &StreamRecorder{}, nil
}

func (sr *StreamRecorder) Start() error {
	sr.cmd = exec.Command("sox",
		"-t", "waveaudio", "-d",
		"-r", "16000",
		"-c", "1",
		"-e", "floating-point",
		"-b", "32",
		"-p",
		"remix", "1",
	)

	var err error
	sr.stdout, err = sr.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	return sr.cmd.Start()
}

func (sr *StreamRecorder) Read(p []byte) (int, error) {
	if sr.stdout == nil {
		return 0, fmt.Errorf("recorder not started")
	}
	return sr.stdout.Read(p)
}

func (sr *StreamRecorder) Stop() error {
	if sr.cmd != nil {
		return sr.cmd.Process.Kill()
	}
	return nil
}

func (sr *StreamRecorder) Close() error {
	if sr.stdout != nil {
		return sr.stdout.Close()
	}
	return nil
}
