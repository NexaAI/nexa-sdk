package record

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
)

type Recorder struct {
	cmd        *exec.Cmd
	outputFile string
}

func tipInstallSox() {
	fmt.Println(render.GetTheme().Warning.Sprintf("sox is not installed. Try:"))
	switch runtime.GOOS {
	case "darwin":
		fmt.Println(render.GetTheme().Warning.Sprintf("  brew install sox"))
	case "linux":
		fmt.Println(render.GetTheme().Warning.Sprintf("  sudo apt install sox       # Debian/Ubuntu"))
		fmt.Println(render.GetTheme().Warning.Sprintf("  sudo yum install sox       # RHEL/CentOS/Fedora"))
	case "windows":
		fmt.Println(render.GetTheme().Warning.Sprintf("  winget install --id=ChrisBagwell.SoX -e"))
	default:
		fmt.Println(render.GetTheme().Warning.Sprintf("sox is not installed. Please install it manually for your OS: %s\n", runtime.GOOS))
	}
}

func NewRecorder(outputFile string) (*Recorder, error) {
	if _, err := exec.LookPath("sox"); err != nil {
		tipInstallSox()
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
