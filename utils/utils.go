package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func IsItRunning(appName string) bool {
	if GetPid(appName) != "0" {
		return true
	}
	return false
}

func GetPid(appName string) string {
	cmd := exec.Command("pidof", appName)
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
	out, err := cmd.Output()
	if err != nil {
		return "0"
	}
	pid := string(out)
	return pid[:len(pid)-1]
}

func GetWindowID(className string) (string, error) {
	cmd := exec.Command("xdotool", "search", "--class", className)
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out[:len(out)-1]), nil
}

func Exec(cmd *exec.Cmd, wait bool) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Ensure DISPLAY is set for all commands
	if cmd.Env == nil {
		cmd.Env = append(os.Environ(), "DISPLAY=:0")
	} else if !hasDisplay(cmd.Env) {
		cmd.Env = append(cmd.Env, "DISPLAY=:0")
	}

	if wait {
		cmd.Run()
	} else {
		cmd.Start()
	}
}

func hasDisplay(env []string) bool {
	for _, e := range env {
		if len(e) >= 8 && e[:8] == "DISPLAY=" {
			return true
		}
	}
	return false
}

func Kill(appName string) {
	killAll := exec.Command("killall", appName)
	killAll.Env = append(os.Environ(), "DISPLAY=:0")
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
