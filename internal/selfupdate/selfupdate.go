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

// Run executes the update. If targetVersion is empty, installs the latest stable.
// If targetVersion is set, always uses the install script (package managers don't support pinning easily).
func Run(method Method, targetVersion string) error {
	if targetVersion != "" {
		return runScriptWithVersion(targetVersion)
	}
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
			"irm https://raw.githubusercontent.com/soverstack/soverstack/main/install.ps1 | iex"))
	default:
		return runInteractive(exec.Command("bash", "-c",
			"curl -fsSL https://raw.githubusercontent.com/soverstack/soverstack/main/install.sh | bash"))
	}
}

func runScriptWithVersion(version string) error {
	version = strings.TrimPrefix(version, "v")
	tag := "v" + version
	fmt.Printf("Installing version %s...\n", version)

	repo := "soverstack/soverstack"
	os_ := runtime.GOOS
	arch := runtime.GOARCH

	switch runtime.GOOS {
	case "windows":
		url := fmt.Sprintf("https://github.com/%s/releases/download/%s/soverstack-%s-%s-%s.zip", repo, tag, version, os_, arch)
		script := fmt.Sprintf(`
$dest = "$env:LOCALAPPDATA\Soverstack\soverstack.exe"
$tmp = "$env:TEMP\soverstack-update.zip"
Invoke-WebRequest -Uri '%s' -OutFile $tmp -UseBasicParsing
Expand-Archive -Path $tmp -DestinationPath "$env:TEMP\soverstack-extract" -Force
Copy-Item "$env:TEMP\soverstack-extract\soverstack.exe" $dest -Force
Remove-Item $tmp -Force -ErrorAction SilentlyContinue
Remove-Item "$env:TEMP\soverstack-extract" -Recurse -Force -ErrorAction SilentlyContinue
Write-Host "soverstack %s installed"
`, url, version)
		return runInteractive(exec.Command("powershell", "-Command", script))

	default:
		ext := "tar.gz"
		url := fmt.Sprintf("https://github.com/%s/releases/download/%s/soverstack-%s-%s-%s.%s", repo, tag, version, os_, arch, ext)
		script := fmt.Sprintf(`
set -e
tmp=$(mktemp -d)
curl -fsSL '%s' -o "$tmp/soverstack.tar.gz"
tar xzf "$tmp/soverstack.tar.gz" -C "$tmp"
if [ -w /usr/local/bin ]; then
  mv "$tmp/soverstack" /usr/local/bin/soverstack
else
  sudo mv "$tmp/soverstack" /usr/local/bin/soverstack
fi
rm -rf "$tmp"
echo "soverstack %s installed"
`, url, version)
		return runInteractive(exec.Command("bash", "-c", script))
	}
}

func runInteractive(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
