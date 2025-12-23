package helper

func ShortenPath(path string, maxWidth int) string {
	if maxWidth <= 0 {
		return  "" 
	}

	w := len(path) 
	if w <= maxWidth {
		return path 
	}

	return "..." + path[w-maxWidth:]
}