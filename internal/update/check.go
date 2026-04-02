package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	releaseURL = "https://api.github.com/repos/soverstack/soverstack/releases/latest"
	timeout    = 3 * time.Second
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckForUpdate checks GitHub for a newer version and returns a message if available.
// Returns empty string if up-to-date or if the check fails (fail silently).
func CheckForUpdate(currentVersion string) string {
	if currentVersion == "dev" || strings.HasSuffix(currentVersion, "-dev") {
		return ""
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(releaseURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if latest == "" || latest == current {
		return ""
	}

	if compareVersions(latest, current) > 0 {
		return fmt.Sprintf(
			"\n  Update available: %s → %s\n"+
				"  Run: soverstack update\n",
			current, latest,
		)
	}

	return ""
}

// compareVersions compares two semver strings (e.g. "1.2.3" vs "1.3.0").
// Returns >0 if a > b, <0 if a < b, 0 if equal.
func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	for i := 0; i < 3; i++ {
		var va, vb int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &va)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &vb)
		}
		if va != vb {
			return va - vb
		}
	}
	return 0
}
