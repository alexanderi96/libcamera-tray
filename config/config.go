package config

import (
	"github.com/alexanderi96/libcamera-tray/types"
	"github.com/alexanderi96/libcamera-tray/utils"
)

const fileName = "config.json"

var Properties types.Configuration

func init() {
	utils.LoadJson(fileName, &Properties)
}
