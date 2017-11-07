//+lbuild experiment

// ! EXPERIMENTAL
package main

import (
	"errors"
	"regexp"
)

type Mango struct {
	tagFinder *regexp.Regexp
}

func (m *Mango) Init() {
	m.tagFinder = regexp.MustCompile(`(?s)\{\{(.*)\}\}`)
}

func (m *Mango) Parse(tmpl string) (out string, err error) {
	tagIndices := m.tagFinder.FindAllStringIndex(tmpl, -1)
	if len(tagIndices) > 0 {
		if tagIndices[0][0] == 0 {
			return "", errors.New("We don't support tags in the outermost layer yet")
		}
		var lastTag = 0
		var lastID = 0
		for _, tagIndex := range tagIndices {
			var nestingLayer = 0
			for i := tagIndex[0]; i > 0; i-- {
				switch tmpl[i] {
				case '>':
					ii, closeTag, err := m.tasteTagToLeft(tmpl, i)
					if err != nil {
						return "", err
					}
					if closeTag {
						nestingLayer++
					} else {
						_, tagID := m.parseTag(tmpl, ii, i)
						if tagID == "" {
							out += tmpl[lastTag:ii] + m.injectID(ii, i)
							lastID++
						} else {
							out += tmpl[lastTag:i]
						}
					}
				case '<':

				}
			}
		}
	}
	return "", nil
}

func (m *Mango) injectID(start int, end int) string {
	return ""
}

func (m *Mango) parseTag(tmpl string, start int, end int) (tagType string, tagID string) {
	var i = start
	for ; i < end; i++ {
		if tmpl[i] == ' ' {
			break
		}
	}
	tagType = tmpl[start:i]
	i = start
	for ; i < (end - 4); i++ {
		if tmpl[i] == ' ' && tmpl[i+1] == 'i' && tmpl[i+2] == 'd' && tmpl[i+3] == '=' {
			tagID = m.extractAttributeContents(tmpl, i+4, end)
		}
	}
	return tagType, tagID
}

func (m *Mango) extractAttributeContents(tmpl string, i int, end int) (contents string) {
	var start = i
	var quoteChar byte = 0 // nolint
	if m.isHTMLQuoteChar(tmpl[i]) {
		i++
		quoteChar = tmpl[i]
	}
	i += 3
	for ; i < end; i++ {
		if quoteChar != 0 {
			if tmpl[i] == quoteChar {
				break
			}
		} else if tmpl[i] == ' ' {
			break
		}
	}
	return tmpl[start:i]
}

func (m *Mango) isHTMLQuoteChar(char byte) bool {
	return char == '\'' || char == '"'
}

func (m *Mango) tasteTagToLeft(tmpl string, index int) (indexOut int, closeTag bool, err error) {
	var foundLeftBrace = false
	for ; index > 0; index-- {
		// What if the / isn't adjacent to the < but has a space instead? Is that even valid?
		if index >= 1 && tmpl[index] == '/' && tmpl[index-1] == '<' {
			closeTag = true
			break
		} else if tmpl[index] == '<' {
			foundLeftBrace = true
		}
	}
	if !foundLeftBrace {
		return index, closeTag, errors.New("The left portion of the tag is missing")
	}
	return index, closeTag, nil
}
