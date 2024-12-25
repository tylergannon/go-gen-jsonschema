package builder

import (
	"errors"
	"io"
	"os/exec"
)

func RunCommand(command string, workDir string, args ...string) (exitCode int, stdout, stderr string, err error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = workDir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return 0, "", "", err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return 0, "", "", err
	}

	if err := cmd.Start(); err != nil {
		return 0, "", "", err
	}

	stdoutBytes, err := io.ReadAll(stdoutPipe)
	if err != nil {
		return 0, "", "", err
	}
	stderrBytes, err := io.ReadAll(stderrPipe)
	if err != nil {
		return 0, "", "", err
	}

	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), string(stdoutBytes), string(stderrBytes), nil
		}
		return 0, string(stdoutBytes), string(stderrBytes), err
	}

	return cmd.ProcessState.ExitCode(), string(stdoutBytes), string(stderrBytes), nil
}
