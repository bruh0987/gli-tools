package text

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

func Copy(value string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard -Value ([Console]::In.ReadToEnd())")
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		for _, name := range []string{"wl-copy", "xclip", "xsel"} {
			if _, err := exec.LookPath(name); err == nil {
				switch name {
				case "xclip":
					cmd = exec.Command("xclip", "-selection", "clipboard")
				case "xsel":
					cmd = exec.Command("xsel", "--clipboard", "--input")
				default:
					cmd = exec.Command(name)
				}
				break
			}
		}
	}
	if cmd == nil {
		return errors.New("no clipboard command found")
	}
	cmd.Stdin = strings.NewReader(value)
	return cmd.Run()
}
