package record

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os/exec"
	"runtime"
)

type StreamRecorder struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func NewStreamRecorder() (*StreamRecorder, error) {
	var args []string

	sr := StreamRecorder{}

	switch runtime.GOOS {
	case "windows":
		args = []string{
			// input (device)
			"-t", "waveaudio", "-d",
			// ----- output format options (apply to the *next* filename) -----
			"-t", "raw", // output container/type
			"-e", "float", // encoding: IEEE float
			"-b", "32", // 32-bit
			"-r", "16000", // output header params
			"-c", "1",
			"-", // OUTFILE = stdout
			// ----- effects (to actually convert from device's 48k/stereo) -----
			"rate", "16000",
			"channels", "1",
		}
	case "darwin", "linux":
		args = []string{
			"-d",
			// input (device)
			"-t", "waveaudio", "-d",
			// ----- output format options (apply to the *next* filename) -----
			"-t", "raw", // output container/type
			"-e", "float", // encoding: IEEE float
			"-b", "32", // 32-bit
			"-r", "16000", // output header params
			"-c", "1",
			"-", // OUTFILE = stdout
			// ----- effects (to actually convert from device's 48k/stereo) -----
			"rate", "16000",
			"channels", "1",
		}
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	sr.cmd = exec.Command("sox", args...)

	var err error
	sr.stdout, err = sr.cmd.StdoutPipe()
	//sr.cmd.Stderr = os.Stderr
	if err != nil {
		return nil, err
	}

	return &sr, nil
}

func (sr *StreamRecorder) Start() error {
	return sr.cmd.Start()
}

func (sr *StreamRecorder) ReadFloat32(buffer []float32) (int, error) {
	if sr.stdout == nil {
		return 0, fmt.Errorf("recorder not started")
	}

	rawBytes := make([]byte, len(buffer)*4)
	n, err := sr.stdout.Read(rawBytes)
	if err != nil {
		return 0, err
	}

	sampleCount := n / 4
	for i := range sampleCount {
		bits := binary.LittleEndian.Uint32(rawBytes[i*4 : (i+1)*4])
		buffer[i] = math.Float32frombits(bits)
	}

	return sampleCount, nil
}

func (sr *StreamRecorder) Stop() error {
	if sr.cmd != nil {
		return sr.cmd.Process.Kill()
	}
	return nil
}
