# libcamera-tray

A libcamera-apps wrapper written in go that provides a convenient ~~system tray icon~~ ui for running the preview and taking a shot.

## Features

- [x] Run the preview
- [x] Take a shot
- [X] See the active libcamera-apps parameters
- [x] Set custom parameters from a json file
- [] Editing the parameters dyrectly from the ui

## HW setup

### Actual setup

![My Frankenstein camera](https://i.imgur.com/XQpxm1D.jpeg)

- Rpi 4 4gb
- CPU-RAM-USB chip heatsinks
- Camera HQ module
- Telephoto lens: [25mm F1.4 CCTV Lens](https://www.amazon.it/gp/product/B00PGOQQ1W/ref=ppx_od_dt_b_asin_title_s00?ie=UTF8&psc=1)
- Wide lens: [6mm lens](https://thepihut.com/products/raspberry-pi-high-quality-camera-lens?variant=31811254190142)
- [Pi zero mounting plate for camera HQ](https://thepihut.com/products/mounting-plate-for-high-quality-camera?variant=31867507048510)
- [Waveshare 4.3inch Capacitive DSI display with case](https://www.waveshare.com/4.3inch-DSI-LCD-with-case.htm)
- Power bank
- Optional mini wireless keyboard + touchpad

### Future ideas

- Phisical shutter button because a camera without it not really comfortable 
- 3D printed custom enclousure based on [Kevin McAleer PIKON project](https://youtu.be/4BEjKUK8DSQ)
- Installing a PWM fan because it is annoying to have it always on


## Screenshots

![Ultrawide screenshot](https://i.imgur.com/7tzBfK3.png)

![Telephoto screenshot](https://i.imgur.com/uR2rzke.png)

[Other images](https://imgur.com/gallery/kcZC4I1)

## Why?

Because I wanted to learn how to program in GO more complex programs and also how to write GUI programs with it. Also I had a Pi4 with a Camera HQ laying around taking dust, so why not.

## Limitations

- Compared to calssic python (or c++) camera implementations, golang has not a native way to interface itself to the camera stack, Therefore the various libcamera-apps must be runned using the exec command.
- Gio is a great golang gui library, but still fairly new. Because of that for example after the window has been created, it must be moved to the correct position using `xdotool` 
