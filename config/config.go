package config

import (
	"os"
	
	"github.com/alexanderi96/libcamera-tray/types"
	"github.com/alexanderi96/libcamera-tray/utils"
)

const fileName = "config.json"

var Properties types.Configuration

func init() {
	// Try to load config file, use defaults if not found
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		Properties = types.Configuration{
			App: types.App{
				Title:  "libcamera-tray",
				Width:  800,
				Height: 480,
			},
			Preview: types.Preview{
				Enabled: true,
				Width:   640,
				Height:  480,
			},
			DateFormat: "2006-01-02",
			TimeFormat: "15:04:05",
		}
		return
	}
	utils.LoadJson(fileName, &Properties)
}
