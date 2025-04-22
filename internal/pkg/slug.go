package pkg

import (
	"fmt"
	"strings"
)

func Slugify(title string, index int) string {
	title = strings.ToLower(title)
	title = strings.ReplaceAll(title, " ", "-")
	title = strings.ReplaceAll(title, ".", "")
	title = strings.ReplaceAll(title, ",", "")
	title = strings.ReplaceAll(title, "!", "")
	title = strings.ReplaceAll(title, "?", "")
	title = strings.ReplaceAll(title, ":", "")
	title = strings.ReplaceAll(title, ";", "")
	title = strings.ReplaceAll(title, "'", "")
	title = strings.ReplaceAll(title, "\"", "")
	title = title + "-" + fmt.Sprintf("%d", index)
	return title
}
