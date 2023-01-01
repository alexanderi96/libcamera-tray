package main

import (
    "time"
    "os"
    "os/exec"
    "log"
    "strconv"
    _ "embed"

    "github.com/getlantern/systray"
)


var (
    //go:embed assets/camera.ico
    icon []byte
)

const (
    appName string = "libcamera-tray"
    appDesc string = "libcamera wrapper wreitten in go"
)

func main() {
    systray.Run(onReady, onExit)
}

func onReady() {

    systray.SetIcon(icon)
    systray.SetTitle(appName)
    systray.SetTooltip(appDesc)

    runPreview := systray.AddMenuItem("Preview", "Preview")
    runShot := systray.AddMenuItem("Shot", "Capture the picture.")
        
    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Quit", "Quit the whole app.")

    prev := exec.Command("libcamera-hello", "-t", "0")

    togglePreview := func() {
        if getPid("libcamera-hello") == "0" {
            log.Print("Starting preview.")
            prev = exec.Command("libcamera-hello", "-t", "0")
            runPreview.SetTitle("Stop Preview")
            prev.Start()
        } else {
            log.Print("Stopping preview.")
            runPreview.SetTitle("Preview")
            prev.Process.Kill()
        }
    }

    killPreviewIfAlive := func() (wasAlive bool) {
        wasAlive = false
        if getPid("libcamera-hello") != "0" {
            wasAlive = true
            togglePreview()
        }
        return
    }

    for {
        select {
        case <-runPreview.ClickedCh:
            togglePreview()
        
        case <-runShot.ClickedCh:
            wasPreviewOpen := killPreviewIfAlive()
            log.Print("Taking a shot.")
            currDate := time.Now()
            epoch := strconv.FormatInt(currDate.Unix(), 10)
            
            homeFolder, err := os.UserHomeDir()
            if err != nil {
                log.Fatal( err )
            }
            
            path := homeFolder + "/Pictures/libcamera-tray/" + currDate.Format("01-02-2006") + "/"

            if err := os.MkdirAll(path, os.ModePerm); err != nil {
                log.Println(err)
            }

            shot := exec.Command("libcamera-still", "-n", "-o", path + "pic" + epoch + ".jpg")
            shot.Run()
            if wasPreviewOpen {
                togglePreview()
            }

        case <-mQuit.ClickedCh:
            _ = killPreviewIfAlive()
            systray.Quit()
            return
        }
    }
}

func onExit() {
    // Cleaning stuff here.
    log.Print("Exiting now...")
}

func getPid(appName string) string {
    out, err := exec.Command("pidof", appName).Output()
    if err != nil {
        return "0"
    }
    return string(out)
}
