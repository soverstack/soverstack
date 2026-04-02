package selfupdate

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Method int

const (
	Homebrew Method = iota
	Scoop
	Script
)

func (m Method) String() string {
	switch m {
	case Homebrew:
		return "homebrew"
	case Scoop:
		return "scoop"
	default:
		return "script"
	}
}

// Detect returns the installation method used on this system.
func Detect() Method {
	switch runtime.GOOS {
	case "windows":
		if isScoop() {
			return Scoop
		}
		return Script
	default:
		if isHomebrew() {
			return Homebrew
		}
		return Script
	}
}

// Run executes the update using the detected installation method.
func Run(method Method) error {
	switch method {
	case Homebrew:
		return runHomebrew()
	case Scoop:
		return runScoop()
	default:
		return runScript()
	}
}

func isHomebrew() bool {
	out, err := exec.Command("brew", "list", "soverstack").CombinedOutput()
	return err == nil && len(out) > 0
}

func isScoop() bool {
	exe, err := exec.LookPath("soverstack.exe")
	if err != nil {
		exe, err = exec.LookPath("soverstack")
	}
	return err == nil && strings.Contains(strings.ToLower(exe), "scoop")
}

func runHomebrew() error {
	fmt.Println("Updating via Homebrew...")
	return runInteractive(exec.Command("brew", "upgrade", "soverstack"))
}

func runScoop() error {
	fmt.Println("Updating via Scoop...")
	return runInteractive(exec.Command("scoop", "update", "soverstack"))
}

func runScript() error {
	fmt.Println("Updating via install script...")
	switch runtime.GOOS {
	case "windows":
		return runInteractive(exec.Command("powershell", "-Command",
			"irm https://raw.githubusercontent.com/soverstack/cli-launcher/main/install.ps1 | iex"))
	default:
		return runInteractive(exec.Command("bash", "-c",
			"curl -fsSL https://raw.githubusercontent.com/soverstack/cli-launcher/main/install.sh | bash"))
	}
}

func runInteractive(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
