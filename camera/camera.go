package camera

import (
    "fmt"
    "time"
    "os"
    "os/exec"
    
    "log"
    "strconv"
    "strings"
    _ "embed"

    "github.com/alexanderi96/libcamera-tray/types"
    "github.com/alexanderi96/libcamera-tray/utils"

)

var (
    //go:embed defaultParams.lctp
    defaultParamsJson []byte

    Params types.ParamsMap

    dateFormat string = "2006-01-02"
    timeFormat string = "15:04:05"

)

func init() {
    Params.LoadParamsMap(defaultParamsJson)

    // I set the custom preview size to fit the waveshare screen
    preview := Params["preview"]
    preview.Value = "0,0,564,423"
    Params["preview"] = preview
}

func TogglePreview() (running bool) {
    running = utils.IsItRunning("libcamera-hello")
    if !running {
        log.Print("Starting preview.")
        prev := buildCommand("libcamera-hello")
        log.Print(prev) 
        utils.Exec(prev, false)
    } else {
        log.Print("Stopping preview.")
        utils.Kill("libcamera-hello")
    }
    return
}

func Shot() {
    running := false
    if running = utils.IsItRunning("libcamera-hello"); running {
        TogglePreview()
    }
    log.Print("Taking a shot.")
    shot := buildCommand("libcamera-still")
    log.Print(shot)
    utils.Exec(shot, true)
    if running {
        TogglePreview()
    }
}

func buildCommand(app string) *exec.Cmd {
    fullString := ""

    if app != "libcamera-hello" {
        fullString = fmt.Sprintf("%s", getOutputPath())
    } else if app == "libcamera-hello" {
        fullString = fmt.Sprintf("%s", "--timeout 0")
    }

    log.Println(fullString)
    
    for key, option := range Params {
        if option.IsCustom() {
            
            switch key {

                case "timestamp", "immediate", "timelapse", "timeout", "framestart", "output":
                    if app != "libcamera-hello" {
                        fullString = fmt.Sprintf("%s --%s %s", fullString, key, option.Value)
                    }

                default:
                    fullString = fmt.Sprintf("%s --%s %s", fullString, key, option.Value)
            }
            
        }
    }
    return exec.Command(app, strings.Split(fullString, " ")...)
}

func getOutputPath() string {

    currDate := time.Now()
    
    folder := ""

    if Params["output"].IsCustom() {
        folder = Params["output"].Value
    } else {
        homeFolder, err := os.UserHomeDir()
        if err != nil {
            log.Fatal( err )
        }
        folder = fmt.Sprintf("%s/%s", homeFolder, "Pictures/libcamera-tray")
    }

    path := fmt.Sprintf("%s/%s", folder, currDate.Format(dateFormat))

    if Params["timelapse"].IsCustom() {
        path = fmt.Sprintf("%s/%s/%s", path, "timelapses", currDate.Format(timeFormat))
    }
    
    if err := os.MkdirAll(path, os.ModePerm); err != nil {
        log.Fatal(err)
    }

    if Params["timelapse"].IsCustom() ||
        Params["datetime"].IsCustom() ||
        Params["timestamp"].IsCustom() {
        return fmt.Sprintf("--output %s", path)
    } else {
        return fmt.Sprintf("--output %s/pic%s.jpg", path, strconv.FormatInt(currDate.Unix(), 10))
    }
}
