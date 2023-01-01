package main

import (
    "fmt"
    "time"
    "os"
    "os/exec"
    "log"
    "strconv"
    "encoding/json"
    _ "embed"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/layout"

)

type Parameter struct {
    Command string
    Default string
    Description string
}

var (
    //go:embed assets/camera.ico
    icon []byte

    //go:embed defaultParams.json
    defaultParamsJson []byte

    defaultParams []Parameter
    customParams []Parameter
)

const (
    appName string = "libcamera-tray"
    appDesc string = "libcamera wrapper wreitten in go"
    saveFolder string = "/Pictures/libcamera-tray/"
)

func main() {
    defaultParams = loadDefaultParams()
    
    a := app.New()
    w := a.NewWindow(appName)

    // let's check thhat we are running on a desktop, so we can show the tray icon
    // if desk, ok := a.(desktop.App); ok {
    //     m := fyne.NewMenu(appName,
    //         fyne.NewMenuItem("Show Preview", func() {togglePreview()}),
    //         fyne.NewMenuItem("Shot", func() {shot()}),
    //         fyne.NewMenuItem("Settings", func() {w.Show()}),
    //     )
    //     desk.SetSystemTrayMenu(m)
    // }

    
    
    previewBtn := widget.NewButton("Toggle Preview", func() {
        togglePreview()
    })
    shotBtn := widget.NewButton("Shot", func() {
        shot()
    })


    tabs := container.NewAppTabs(
        container.NewTabItemWithIcon("Home", theme.HomeIcon(), container.New(layout.NewVBoxLayout(), previewBtn, shotBtn)),
        container.NewTabItem("Settings", widget.NewList(
        func() int {
            return len(defaultParams)
        },
        func() fyne.CanvasObject {
            return widget.NewLabel("template")
        },
        func(i widget.ListItemID, o fyne.CanvasObject) {
            o.(*widget.Label).SetText(fmt.Sprintf("%s = %s", defaultParams[i].Command, defaultParams[i].Default))
        })),
    )

    tabs.SetTabLocation(container.TabLocationLeading)

    w.SetContent(tabs)
    //w.SetCloseIntercept(func() {w.Hide()})
    w.Resize(fyne.Size{300, 200})
    w.ShowAndRun()
}

func loadDefaultParams() (parametri []Parameter) {
    err := json.Unmarshal(defaultParamsJson, &parametri)

    if err != nil {
        log.Fatal(err)
    }

    return
}

func getSettings() {
    //....
}

func togglePreview() {
    if !isItRunning("libcamera-hello") {
        log.Print("Starting preview.")
        prev := exec.Command("libcamera-hello", "-t", "0")
        prev.Start()
    } else {
        log.Print("Stopping preview.")

        pid, err := strconv.Atoi(getPid("libcamera-hello"))

        if err != nil {
            log.Fatal(err)
        }

        proc, err := os.FindProcess(pid)

        if err != nil {
            log.Fatal(err)
        }

        proc.Kill()
    }
}

func shot() {
    running := false
    if running = isItRunning("libcamera-hello"); running {
        togglePreview()
    }
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

    shot := exec.Command("libcamera-still", "-n", "--immediate", "-o", path + "pic" + epoch + ".jpg")
    shot.Run()
    if running {
        togglePreview()
    }
}            

func isItRunning(appName string) bool {
    if getPid(appName) != "0" {
        return true
    }
    return false
}

func getPid(appName string) string {
    out, err := exec.Command("pidof", appName).Output()
    if err != nil {
        return "0"
    }
    pid := string(out)
    return pid[:len(pid)-1]
}
