package tmpl

import (
	"strconv"
	"strings"
)

// TODO: Write unit tests for this
func minify(data string) string {
	data = strings.Replace(data, "\t", "", -1)
	data = strings.Replace(data, "\v", "", -1)
	data = strings.Replace(data, "\n", "", -1)
	data = strings.Replace(data, "\r", "", -1)
	data = strings.Replace(data, "  ", " ", -1)
	return data
}

// TODO: Strip comments
// TODO: Handle CSS nested in <style> tags?
// TODO: Write unit tests for this
func minifyHTML(data string) string {
	return minify(data)
}

// TODO: Have static files use this
// TODO: Strip comments
// TODO: Convert the rgb()s to hex codes?
// TODO: Write unit tests for this
func minifyCSS(data string) string {
	return minify(data)
}

// TODO: Convert this to three character hex strings whenever possible?
// TODO: Write unit tests for this
// nolint
func rgbToHexstr(red int, green int, blue int) string {
	return strconv.FormatInt(int64(red), 16) + strconv.FormatInt(int64(green), 16) + strconv.FormatInt(int64(blue), 16)
}

/*
// TODO: Write unit tests for this
func hexstrToRgb(hexstr string) (red int, blue int, green int, err error) {
	// Strip the # at the start
	if hexstr[0] == '#' {
		hexstr = strings.TrimPrefix(hexstr,"#")
	}
	if len(hexstr) != 3 && len(hexstr) != 6 {
		return 0, 0, 0, errors.New("Hex colour codes may only be three or six characters long")
	}

	if len(hexstr) == 3 {
		hexstr = hexstr[0] + hexstr[0] + hexstr[1] + hexstr[1] + hexstr[2] + hexstr[2]
	}
}*/
