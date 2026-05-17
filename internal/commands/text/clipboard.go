package text

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Copy(value string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		return copyWindows(value)
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

func copyWindows(value string) error {
	file, err := os.CreateTemp("", "gli-clipboard-*.txt")
	if err != nil {
		return err
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.WriteString(value); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	script := fmt.Sprintf("Set-Clipboard -Value (Get-Content -Raw -LiteralPath '%s')", strings.ReplaceAll(path, "'", "''"))
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	return cmd.Run()
}
