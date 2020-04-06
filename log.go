package main

var (
	colorGreenBg   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	colorWhiteBg   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	colorYellowBg  = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	colorRedBg     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	colorBlueBg    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	colorMagentaBg = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	colorCyanBg    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})

	colorGreen   = string([]byte{27, 91, 51, 50, 109})
	colorWhite   = string([]byte{27, 91, 51, 55, 109})
	colorYellow  = string([]byte{27, 91, 51, 51, 109})
	colorRed     = string([]byte{27, 91, 51, 49, 109})
	colorBlue    = string([]byte{27, 91, 51, 52, 109})
	colorMagenta = string([]byte{27, 91, 51, 53, 109})
	colorCyan    = string([]byte{27, 91, 51, 54, 109})

	colorReset = string([]byte{27, 91, 48, 109})
)

func logSuccess(text string, args ...interface{}) {
	args = append([]interface{}{colorGreen}, args...)
	args = append(args, colorReset)
	logger.Printf("%s"+text+"%s", args...)
}

func logInfo(text string, args ...interface{}) {
	args = append([]interface{}{colorCyan}, args...)
	args = append(args, colorReset)
	logger.Printf("%s"+text+"%s", args...)
}

func logError(text string, args ...interface{}) {
	args = append([]interface{}{colorRed}, args...)
	args = append(args, colorReset)
	logger.Printf("%s"+text+"%s", args...)
}

func logWarn(text string, args ...interface{}) {
	args = append([]interface{}{colorYellow}, args...)
	args = append(args, colorReset)
	logger.Printf("%s"+text+"%s", args...)
}

func logDebug(text string, args ...interface{}) {
	args = append([]interface{}{colorMagenta}, args...)
	args = append(args, colorReset)
	logger.Printf("%s"+text+"%s", args...)
}
