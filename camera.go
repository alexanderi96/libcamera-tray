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
    bot string = "/workspace/myrepo/PiCamBot/PiCamBot"
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
    runBot := systray.AddMenuItem("Run Bot", "Bot")

    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Quit", "Quit the whole app.")

    prev := exec.Command("libcamera-hello", "-t", "0")

    homeFolder, err := os.UserHomeDir()
    if err != nil {
        log.Fatal( err )
    }

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

    botPos := homeFolder + bot
    botComm := exec.Command(botPos)

    toggleBot := func() {
        if getPid(botPos) == "0" {
            log.Print("Starting bot.")
            botComm = exec.Command(botPos)
            runBot.SetTitle("Stop Bot")
            botComm.Start()
        } else {
            log.Print("Stopping bot.")
            runBot.SetTitle("Run Bot")
            botComm.Process.Kill()
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
            
            path := homeFolder + "/Pictures/libcamera-tray/" + currDate.Format("01-02-2006") + "/"

            if err := os.MkdirAll(path, os.ModePerm); err != nil {
                log.Println(err)
            }

            shot := exec.Command("libcamera-still", "-n", "-o", path + "pic" + epoch + ".jpg")
            shot.Run()
            if wasPreviewOpen {
                togglePreview()
            }

        case <-runBot.ClickedCh:
            toggleBot()

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
