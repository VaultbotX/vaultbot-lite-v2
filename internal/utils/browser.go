package utils

import (
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"os/exec"
	"runtime"
)

func OpenBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = domain.ErrUnsupportedOSForBrowser
	}

	return err
}
