package main

import (
    "fmt"
    "time"
    "os"
    "os/exec"
    "log"
    "io/ioutil"
    "strconv"
    "strings"
    "encoding/json"
    _ "embed"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"

    "fyne.io/fyne/v2/storage"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/layout"

)

type Parameter struct {
    Command string `json:command`
    Value string `json:value`
    Enabled bool `json:enabled`
    Description string `json:description`
}

var (
    //go:embed assets/camera.ico
    icon []byte

    //go:embed defaultParams.lctp
    defaultParamsJson []byte

    defaultParams = loadParamsMap(defaultParamsJson)
    customParams = make(map[string]Parameter)

    defaultParamsKeyList = generateKeyList(defaultParams)
    customParamsKeyList []string
)

const (
    appName string = "libcamera-tray"
    appDesc string = "libcamera-apps wrapper wreitten in go"
    dateFormat string = "2006-01-02"
)

func main() {
    a := app.New()
    w := a.NewWindow(appName)

    //let's check thhat we are running on a desktop, so we can show the tray icon
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
        container.NewTabItem("Settings", 
            container.NewBorder(
                //sopra select con tutte le opzioni standard e bottone per aggiungerla a quelle selezionate
                widget.NewSelect(
                    defaultParamsKeyList,
                    nil,
                    
                ),
                //sotto i vari bottoni per caricare o salvare le impostazioni
                widget.NewButton("Load presets", func() {
                    fileDialog := dialog.NewFileOpen(
                        func(r fyne.URIReadCloser, _ error) {
                            // read files
                            data, _ := ioutil.ReadAll(r)
                            
                            customParams = loadParamsMap(data)
                            customParamsKeyList = generateKeyList(customParams)
                            if isItRunning("libcamera-hello") {
                                togglePreview()
                                togglePreview()
                            }
                        },
                    w)
                    // fiter to open .txt files only
                    // array/slice of strings/extensions
                    fileDialog.SetFilter(
                        storage.NewExtensionFileFilter([]string{".lctp"}))

                    fileDialog.Show()
                    // Show file selection dialog.
                    },
                ),
                nil,
                nil,
                //al centro la lista di opzioni selezionate
                widget.NewList(
                    func() int {
                        return len(customParams)
                    },
                    func() fyne.CanvasObject {
                        return widget.NewLabel("Selected Options")
                    },
                    func(i widget.ListItemID, o fyne.CanvasObject) {
                        o.(*widget.Label).SetText(fmt.Sprintf("%s %s", customParamsKeyList[i], customParams[customParamsKeyList[i]].Value))
                    },
                ),
            ),
        ),

    )

    tabs.SetTabLocation(container.TabLocationLeading)

    w.SetContent(tabs)
    //w.SetCloseIntercept(func() {w.Hide()})
    w.Resize(fyne.Size{600, 400})
    w.ShowAndRun()
}


func loadParams(bytes []byte) (params []Parameter) {
    err := json.Unmarshal(bytes, &params)

    if err != nil {
        log.Fatal(err)
    }

    return
}

func loadParamsMap(bytes []byte) (params map[string]Parameter) {
    err := json.Unmarshal(bytes, &params)

    if err != nil {
        log.Fatal(err)
    }

    return
}

func generateKeyList(params map[string]Parameter) (list []string) {
    for key := range params {
        list = append(list, key)
    }
    return
}

func togglePreview() {
    if !isItRunning("libcamera-hello") {
        log.Print("Starting preview.")
        prev := exec.Command("libcamera-hello", buildCommand("libcamera-hello", "--timeout 0")...)
        log.Print(prev) 
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
    shot := exec.Command("libcamera-still", buildCommand("libcamera-still", "")...)
    log.Print(shot)    
    shot.Run()
    if running {
        togglePreview()
    }
}

func buildCommand(app, baseCommand string) (out []string) {
    fullString := baseCommand
    for key, content := range customParams {
        if (content.Value != defaultParams[key].Value) {
            
            switch key {
                case "output":
                    homeFolder, err := os.UserHomeDir()
                    if err != nil {
                        log.Fatal( err )
                    }

                    currDate := time.Now()

                    path := fmt.Sprintf("%s/%s/%s", homeFolder, content.Value, currDate.Format(dateFormat))
                    
                    if err := os.MkdirAll(path, os.ModePerm); err != nil {
                        log.Fatal(err)
                    }

                    if customParams["datetime"] != defaultParams["datetime"] || customParams["timestamp"] != defaultParams["timestamp"] {
                        fullString = fmt.Sprintf("%s --%s %s", fullString, key, path)
                    } else {
                        fullString = fmt.Sprintf("%s --%s %s/pic%s.jpg", fullString, key, path, strconv.FormatInt(currDate.Unix(), 10))
                    }

                case "timestamp":
                case "immediate":
                    if app != "libcamera-hello" {
                        fullString = fmt.Sprintf("%s --%s %s", fullString, key, content.Value)
                    }
                    
                default:
                    fullString = fmt.Sprintf("%s --%s %s", fullString, key, content.Value)
            }
            
        }
    }
    return strings.Split(fullString, " ")
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
