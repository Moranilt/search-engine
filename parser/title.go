package parser

func ExtractTitle(html []byte) string {

	for i, char := range html {
		if char == '<' && string(html[i+1:i+6]) == "title" {
			j := i + 5
			for html[j] != '<' {
				j++
			}
			return string(html[i+7 : j])
		}
	}

	return ""
}
