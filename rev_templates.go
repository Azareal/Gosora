//+build experiment

// ! EXPERIMENTAL
package main

import (
	"errors"
	"regexp"
)

var tagFinder *regexp.Regexp

func init() {
	tagFinder = regexp.MustCompile(`(?s)\{\{(.*)\}\}`)
}

func icecreamSoup(tmpl string) error {
	tagIndices := tagFinder.FindAllStringIndex(tmpl, -1)
	if tagIndices != nil && len(tagIndices) > 0 {

		if tagIndices[0][0] == 0 {
			return errors.New("We don't support tags in the outermost layer yet")
		}
		for _, tagIndex := range tagIndices {
			var nestingLayer = 0
			for i := tagIndex[0]; i > 0; i-- {
				switch tmpl[i] {
				case '>':
					i, closeTag, err := tasteTagToLeft(tmpl, i)
					if err != nil {
						return err
					}
					if closeTag {
						nestingLayer++
					}
				case '<':

				}
			}
		}
	}
}

func tasteTagToLeft(tmpl string, index int) (indexOut int, closeTag bool, err error) {
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
		return errors.New("The left portion of the tag is missing")
	}
	return index, closeTag, nil
}
