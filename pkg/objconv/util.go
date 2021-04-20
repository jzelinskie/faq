package objconv

import "os"

// trueColorSupported returns true if the tty is configured to support
// truecolor.
func trueColorSupported() bool {
	return os.Getenv("COLORTERM") == "truecolor"
}

// ChromaFormatter is a helper to detect the ideal Chroma formatter name for
// colorizing output.
//
// This function is useful for implementing Color() in the Encoding interface.
func ChromaFormatter() string {
	formatter := os.Getenv("FAQ_FORMATTER")
	if formatter == "" {
		formatter = "terminal"
	} else if trueColorSupported() {
		formatter = "terminal16m"
	}
	return formatter
}

// ChromaStyle is a helper to return the default Chroma style.
//
// This function is useful for implementing Color() in the Encoding interface.
func ChromaStyle() string {
	style := os.Getenv("FAQ_STYLE")
	if style == "" {
		return "pygments"
	}
	return style
}
