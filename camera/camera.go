package camera

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"
	"sync"
	"image"
	_ "embed"
	"strconv"
	"strings"
	"io"

	"github.com/alexanderi96/libcamera-tray/config"
	"github.com/alexanderi96/libcamera-tray/types"
	"github.com/alexanderi96/libcamera-tray/utils"
)

var (
	//go:embed defaultParams.lctp
	defaultParamsJson []byte

	DefaultParams types.ParamsMap
	Params       types.ParamsMap
	homeFolder   string
	
	previewMutex sync.RWMutex
	previewCmd   *exec.Cmd
	previewChan  chan image.Image
	stopPreview  chan struct{}
	
	// Statistics
	frameStats struct {
		sync.Mutex
		received uint64
		dropped  uint64
		errors   uint64
		lastFPS  float64
	}
)

// GetFrameStats returns current frame processing statistics
func GetFrameStats() (received, dropped, errors uint64, fps float64) {
	frameStats.Lock()
	defer frameStats.Unlock()
	return frameStats.received, frameStats.dropped, frameStats.errors, frameStats.lastFPS
}

func init() {
	var err error
	homeFolder, err = os.UserHomeDir()
	if err != nil {
		utils.Error("Failed to get user home directory: %v", err)
		os.Exit(1)
	}

	DefaultParams.LoadParamsMap(defaultParamsJson)

	if config.Properties.ConfigPath != "" {
		Params.LoadParamsMap(utils.OpenFile(fmt.Sprintf("%s/%s", homeFolder, config.Properties.ConfigPath)))
	} else {
		Params.LoadParamsMap(defaultParamsJson)
	}

	// Set the preview size to match the configured dimensions
	if preview, ok := Params["preview"]; ok {
		preview.Enabled = true
		Params["preview"] = preview
	}
}

func IsPreviewRunning() bool {
	return utils.IsItRunning("libcamera-vid")
}

func TogglePreview() (running bool) {
	running = IsPreviewRunning()
	if !running {
		StartPreview()
	} else {
		StopPreview()
	}
	return
}

func StartPreview() chan image.Image {
	previewMutex.Lock()
	defer previewMutex.Unlock()

	if previewChan != nil {
		return previewChan
	}

	utils.Info("Starting preview")
	utils.Debug("Creating preview channels")
	previewChan = make(chan image.Image, 1) // Minimal buffer to prevent accumulation
	stopPreview = make(chan struct{})

	go func() {
		defer close(previewChan)
		
		// For preview, we don't want to use buildCommand as it adds output path
		args := []string{
			"--codec", "mjpeg",
			"--inline",
			"--output", "-",
			"--verbose", "2",
			"--width", strconv.Itoa(config.Properties.Preview.Width),
			"--height", strconv.Itoa(config.Properties.Preview.Height),
			"--framerate", "30", // Increased for smoother preview
			"--denoise", "cdn_off",
			"--nopreview",
			"--timeout", "0", // Run indefinitely
		}
		
		utils.Info("Creating libcamera-vid command with args: %v", args)
		prev := exec.Command("libcamera-vid", args...)
		
		// Check if libcamera-vid exists
		if _, err := exec.LookPath("libcamera-vid"); err != nil {
			utils.Error("libcamera-vid not found: %v", err)
			return
		}
		
		utils.Debug("Setting up stdout pipe")
		stdout, err := prev.StdoutPipe()
		if err != nil {
			utils.Error("Failed to create stdout pipe: %v", err)
			return
		}

		utils.Debug("Setting up stderr pipe for logging")
		stderr, err := prev.StderrPipe()
		if err != nil {
			utils.Error("Failed to create stderr pipe: %v", err)
			return
		}

		// Start stderr logging goroutine
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				utils.Info("libcamera-vid: %s", scanner.Text())
			}
		}()

		utils.Debug("Starting libcamera-vid process")
		if err := prev.Start(); err != nil {
			utils.Error("Failed to start preview: %v", err)
			return
		}
		utils.Info("libcamera-vid process started with PID: %d", prev.Process.Pid)

		previewCmd = prev
		utils.Info("Preview process started successfully")
		
		go func() {
			<-stopPreview
			if previewCmd != nil && previewCmd.Process != nil {
				previewCmd.Process.Kill()
			}
		}()

		reader := NewMJPEGFrameReader(stdout)
		frameCount := 0
		lastLog := time.Now()

		for {
			select {
			case <-stopPreview:
				return
			default:
				img, err := reader.ReadFrame()
				if err != nil {
					if err != io.EOF {
						utils.Error("Error reading frame: %v", err)
					}
					return
				}
				
				frameCount++
				now := time.Now()
				if now.Sub(lastLog) >= time.Second {
					fps := float64(frameCount) / now.Sub(lastLog).Seconds()
					
					frameStats.Lock()
					frameStats.lastFPS = fps
					frameStats.received += uint64(frameCount)
					frameStats.Unlock()
					
					utils.Info("Preview stats - FPS: %.2f, Total Received: %d, Dropped: %d, Errors: %d",
						fps, frameStats.received, frameStats.dropped, frameStats.errors)
					
					frameCount = 0
					lastLog = now
				}

				// Check if we should process this frame
				select {
				case <-stopPreview:
					return
				default:
					// Try to send frame, but drop immediately if channel is busy
					select {
					case previewChan <- img:
						// Frame sent successfully
					default:
						// Drop frame immediately if we can't send it
						frameStats.Lock()
						frameStats.dropped++
						frameStats.Unlock()
						utils.Debug("Frame dropped - channel busy")
					}
				}
			}
		}
	}()

	return previewChan
}

func StopPreview() {
	previewMutex.Lock()
	defer previewMutex.Unlock()

	utils.Debug("Stopping preview")
	if stopPreview != nil {
		close(stopPreview)
		stopPreview = nil
	}

	if previewCmd != nil && previewCmd.Process != nil {
		utils.Debug("Killing preview process")
		previewCmd.Process.Kill()
		previewCmd = nil
	}

	// Drain any remaining frames from the channel
	if previewChan != nil {
		utils.Debug("Draining preview channel")
		for {
			select {
			case _, ok := <-previewChan:
				if !ok {
					goto done
				}
			default:
				goto done
			}
		}
	done:
		previewChan = nil
	}

	utils.Info("Preview stopped")
}

func StopPreviewAndReload(middle func()) {
	running := false
	if running = utils.IsItRunning("libcamera-vid"); running {
		utils.Debug("Stopping preview for reload")
		StopPreview() // Use StopPreview directly to ensure proper cleanup
	}
	if middle != nil {
		middle()
	}
	if running {
		utils.Debug("Restarting preview after reload")
		StartPreview() // Start fresh preview
	}
}

func Shot() chan struct{} {
	done := make(chan struct{})
	
	go func() {
		defer close(done)
		utils.Info("Taking a shot")
		// Stop preview and clean up
		if utils.IsItRunning("libcamera-vid") {
			utils.Debug("Stopping preview for shot")
			StopPreview()
		}

		// Take the shot
		shot := buildCommand("libcamera-still")
		shot.Args = append(shot.Args, "--nopreview")
		utils.Debug("Shot command: %v", shot)
		utils.Exec(shot, true)

		// Signal completion
		utils.Debug("Shot completed")
	}()
	
	return done
}

func buildCommand(app string) *exec.Cmd {
	fullString := ""

	if app != "libcamera-hello" {
		fullString = fmt.Sprintf("%s", getOutputPath())
	} else if app == "libcamera-hello" {
		fullString = fmt.Sprintf("%s", "--timeout 0")
	}

	for key, option := range Params {
		if key != "output" && option.Enabled && option.Value != "" && option.Value != DefaultParams[key].Value {
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

	if Params["output"].Enabled && Params["output"].Value != "" && Params["output"].Value != DefaultParams["output"].Value {
		folder = fmt.Sprintf("%s/%s", homeFolder, Params["output"].Value)
	} else {
		folder = fmt.Sprintf("%s/Pictures/libcamera-tray", homeFolder)
	}

	path := fmt.Sprintf("%s/%s", folder, currDate.Format(config.Properties.DateFormat))

	if Params["timelapse"].Enabled && Params["timelapse"].Value != "" && Params["timelapse"].Value != DefaultParams["timelapse"].Value {
		path = fmt.Sprintf("%s/%s/%s", path, "timelapses", currDate.Format(config.Properties.TimeFormat))
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		utils.Error("Failed to create directory: %v", err)
		os.Exit(1)
	}

	if (Params["timelapse"].Enabled && Params["timelapse"].Value != "" && Params["timelapse"].Value != DefaultParams["timelapse"].Value) ||
		(Params["datetime"].Enabled && Params["datetime"].Value != "" && Params["datetime"].Value != DefaultParams["datetime"].Value) ||
		(Params["timestamp"].Enabled && Params["timestamp"].Value != "" && Params["timestamp"].Value != DefaultParams["timestamp"].Value) {
		return fmt.Sprintf("--output %s", path)
	} else {
		return fmt.Sprintf("--output %s/pic%s.jpg", path, strconv.FormatInt(currDate.Unix(), 10))
	}
}
