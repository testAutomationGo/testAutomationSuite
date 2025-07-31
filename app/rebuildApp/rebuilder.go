package rebuildApp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"testAutomationSuiteGO/app/shared"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// To rebuild the rebuild app app, you need to change this file to a main file temporarily, and RebuilderMain() to a main() function. Then comment out the rebuild app button functionality in the settings menu.
// After rebuilding, you can name back to RebuilderMain() the package back to rebuildApp and uncomment the rebuild app button functionality in the settings menu.

func RebuildDesktopApp() string {
	originalDir, _ := os.Getwd()
	err := os.Chdir(originalDir)
	if err != nil {
		log.Println("Warning: Failed to change directory:", err)
		return "Failed to change directory."
	}
	cmd := exec.Command("go", "build", "-ldflags", "-H=windowsgui", "-o", "TestingApp.exe", "app/appMain.go")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Fatal: Failed to build desktop app: %v\n", err)
		log.Fatalf("Failed to build the desktop app: %v\n", err)
		return "Failed to build desktop app."
	}
	err = os.Chdir(originalDir)
	if err != nil {
		fmt.Printf("Warning: Failed to restore original directory: %v\n", err)
		log.Fatalf("Failed to restore original directory: %v\n", err)
		return "Failed to restore original directory."
	}
	return "Desktop app rebuilt successfully."
}

func ShutDownNotification(deps shared.AppDependencies) {
	myApp := deps.App
	myWindow := myApp.NewWindow("Notification")

	countdownLabel := widget.NewLabel("Closing App in")
	secondsLabel := widget.NewLabel("3 Seconds")

	countdownLabel.Alignment = fyne.TextAlignCenter
	secondsLabel.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		widget.NewSeparator(),
		countdownLabel,
		secondsLabel,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(300, 200))
	myWindow.CenterOnScreen()

	countdown := 1
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		defer ticker.Stop()
		for range ticker.C {
			countdown--

			if countdown > 1 {
				secondsLabel.SetText(fmt.Sprintf("%d Seconds", countdown))
			} else if countdown == 1 {
				secondsLabel.SetText("1 Second")
			} else {
				myApp.Quit()
				return
			}
		}
	}()

	myWindow.ShowAndRun()
}
