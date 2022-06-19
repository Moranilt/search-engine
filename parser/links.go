package parser

import "strings"

func ExtractLinks(html []byte) []string {
	var startIndex int
	var links []string // parse attributes from tags 'a'

	for index, char := range html {
		if char == '<' && html[index+1] == 'a' && html[index+2] == ' ' {
			startIndex = index + 2
		} else if char == '>' && startIndex > 0 {
			links = append(links, string(html[startIndex:index]))
			startIndex = 0
		}
	}

	uniqueLinks := make(map[string]bool)
	var parsedData []string
	for _, attributes := range links {
		link := findLink(attributes)
		if link != "" && !uniqueLinks[link] && link != "#" && strings.Index(link, "https:") == -1 && strings.Index(link, "http:") == -1 {
			uniqueLinks[link] = true
			parsedData = append(parsedData, link)
		}
	}

	return parsedData
}

func findLink(attributes string) string {
	var startIndex int
	for i, char := range attributes {
		if char == ' ' {
			startIndex = i + 1
		} else if char == '=' && attributes[startIndex:i] == "href" {
			hrefIndex := i + 2
			var link []byte
			for attributes[hrefIndex] != attributes[i+1] {
				link = append(link, attributes[hrefIndex])
				hrefIndex++
			}
			return string(link)
		}
	}
	return ""
}
