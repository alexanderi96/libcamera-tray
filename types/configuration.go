package types

type App struct {
	Title       string `json:title`
	Description string `json:description`
	Width       int    `json:width`
	Height      int    `json:height`
	X           int    `json:x`
	Y           int    `json:y`
}

type Preview struct {
	Enabled bool `json:enabled`
	Width   int  `json:width`
	Height  int  `json:height`
	X       int  `json:x`
	Y       int  `json:y`
}

type Configuration struct {
	App        App     `json:app`
	Preview    Preview `json:preview`
	DateFormat string  `json:dateFormat`
	TimeFormat string  `json:timeFormat`
}
