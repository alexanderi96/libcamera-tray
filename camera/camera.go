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
    "github.com/alexanderi96/libcamera-tray/config"

)

var (
    //go:embed defaultParams.lctp
    defaultParamsJson []byte

    DefaultParams types.ParamsMap
    CustomParams types.ParamsMap
)

func init() {
    DefaultParams.LoadParamsMap(defaultParamsJson)
    CustomParams = make(types.ParamsMap)

    // I set the custom preview size to fit the waveshare screen
    preview := CustomParams["preview"]
    preview.Value = fmt.Sprintf("%d,%d,%d,%d",
        config.Properties.Preview.X,
        config.Properties.Preview.Y,
        config.Properties.Preview.Width,
        config.Properties.Preview.Height,
    )
    CustomParams["preview"] = preview
}

func TogglePreview() (running bool) {
    running = utils.IsItRunning("libcamera-hello")
    if !running {
        StartPreview()
    } else {
        StopPreview()
    }
    return
}

func StartPreview() {
    log.Println("Starting preview.")
    prev := buildCommand("libcamera-hello")
    log.Print(prev)
    utils.Exec(prev, false)
}

func StopPreview() {
    log.Println("Stopping preview.")
    utils.Kill("libcamera-hello")
}

func StopPreviewAndReload(middle func()) {
    running := false
    if running = utils.IsItRunning("libcamera-hello"); running {
        TogglePreview()
    }
    if middle != nil {
        middle()
    }
    if running {
        TogglePreview()
    }
}

func Shot() {
    StopPreviewAndReload(func() {
        log.Println("Taking a shot.")
        shot := buildCommand("libcamera-still")
        log.Print(shot)
        utils.Exec(shot, true)
    })
}

func buildCommand(app string) *exec.Cmd {
    fullString := ""

    if app != "libcamera-hello" {
        fullString = fmt.Sprintf("%s", getOutputPath())
    } else if app == "libcamera-hello" {
        fullString = fmt.Sprintf("%s", "--timeout 0")
    }

    for key, option := range CustomParams {
        if option.Value != "" && option.Value != DefaultParams[key].Value{
            switch key {

                case "timestamp", "immediate", "timelapse", "timeout", "framestart", "output", "shutter":
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

    if CustomParams["output"].Value != "" && CustomParams["output"].Value != DefaultParams["output"].Value {
        folder = CustomParams["output"].Value
    } else {
        homeFolder, err := os.UserHomeDir()
        if err != nil {
            log.Fatal( err )
        }
        folder = fmt.Sprintf("%s/%s", homeFolder, "Pictures/libcamera-tray")
    }

    path := fmt.Sprintf("%s/%s", folder, currDate.Format(config.Properties.DateFormat))

    if (CustomParams["timelapse"].Value != "" && CustomParams["timelapse"].Value != DefaultParams["timelapse"].Value) {
        path = fmt.Sprintf("%s/%s/%s", path, "timelapses", currDate.Format(config.Properties.TimeFormat))
    }
    
    if err := os.MkdirAll(path, os.ModePerm); err != nil {
        log.Fatal(err)
    }

    if (CustomParams["timelapse"].Value != "" && CustomParams["timelapse"].Value != DefaultParams["timelapse"].Value) ||
    (CustomParams["datetime"].Value != "" && CustomParams["datetime"].Value != DefaultParams["datetime"].Value) ||
    (CustomParams["timestamp"].Value != "" && CustomParams["timestamp"].Value != DefaultParams["timestamp"].Value) {
        return fmt.Sprintf("--output %s", path)
    } else {
        return fmt.Sprintf("--output %s/pic%s.jpg", path, strconv.FormatInt(currDate.Unix(), 10))
    }
}
