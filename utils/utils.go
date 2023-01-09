package utils

import (
	"os/exec"
	"os"
    "log"
    "io/ioutil"
    "encoding/json"
)

func IsItRunning(appName string) bool {
    if GetPid(appName) != "0" {
        return true
    }
    return false
}

func GetPid(appName string) string {
    out, err := exec.Command("pidof", appName).Output()
    if err != nil {
        return "0"
    }
    pid := string(out)
    return pid[:len(pid)-1]
}

func Exec(cmd *exec.Cmd, wait bool) {
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if wait {
        cmd.Run()    
    } else {
        cmd.Start()
    }
    
}

func Kill(appName string) {
    killAll := exec.Command("killall", appName)
    Exec(killAll, true)
}

func OpenFile(filename string) []byte {
    f, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }

    data, err := ioutil.ReadAll(f)
    if err != nil {
      log.Fatal(err)
    }

    return data
}

func LoadJson[T any](fileName string, obj *T) {
    configFile, err := ioutil.ReadFile(fileName)
    if err != nil {
        log.Fatal("Unable to read file: ", err)
    }

    err = json.Unmarshal(configFile, &obj)
    if err != nil {
        log.Fatal("Invalid file: ", err)
    }
}