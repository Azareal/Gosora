/*
*
*	Query Generator Library
*	WIP Under Construction
*	Copyright Azareal 2017 - 2020
*
 */
package qgen

import (
	"os"
	"strings"
)

// TODO: Add support for numbers and strings?
func processColumns(colstr string) (columns []DBColumn) {
	if colstr == "" {
		return columns
	}
	colstr = strings.Replace(colstr, " as ", " AS ", -1)
	for _, segment := range strings.Split(colstr, ",") {
		var outcol DBColumn
		dotHalves := strings.Split(strings.TrimSpace(segment), ".")

		var halves []string
		if len(dotHalves) == 2 {
			outcol.Table = dotHalves[0]
			halves = strings.Split(dotHalves[1], " AS ")
		} else {
			halves = strings.Split(dotHalves[0], " AS ")
		}

		halves[0] = strings.TrimSpace(halves[0])
		if len(halves) == 2 {
			outcol.Alias = strings.TrimSpace(halves[1])
		}
		if halves[0][len(halves[0])-1] == ')' {
			outcol.Type = "function"
		} else if halves[0] == "?" {
			outcol.Type = "substitute"
		} else {
			outcol.Type = "column"
		}

		outcol.Left = halves[0]
		columns = append(columns, outcol)
	}
	return columns
}

// TODO: Allow order by statements without a direction
func processOrderby(orderstr string) (order []DBOrder) {
	if orderstr == "" {
		return order
	}
	for _, segment := range strings.Split(orderstr, ",") {
		var outorder DBOrder
		halves := strings.Split(strings.TrimSpace(segment), " ")
		if len(halves) != 2 {
			continue
		}
		outorder.Column = halves[0]
		outorder.Order = strings.ToLower(halves[1])
		order = append(order, outorder)
	}
	return order
}

func processJoiner(joinstr string) (joiner []DBJoiner) {
	if joinstr == "" {
		return joiner
	}
	joinstr = strings.Replace(joinstr, " on ", " ON ", -1)
	joinstr = strings.Replace(joinstr, " and ", " AND ", -1)
	for _, segment := range strings.Split(joinstr, " AND ") {
		var outjoin DBJoiner
		var parseOffset int
		var left, right string

		left, parseOffset = getIdentifier(segment, parseOffset)
		outjoin.Operator, parseOffset = getOperator(segment, parseOffset+1)
		right, parseOffset = getIdentifier(segment, parseOffset+1)

		leftColumn := strings.Split(left, ".")
		rightColumn := strings.Split(right, ".")
		outjoin.LeftTable = strings.TrimSpace(leftColumn[0])
		outjoin.RightTable = strings.TrimSpace(rightColumn[0])
		outjoin.LeftColumn = strings.TrimSpace(leftColumn[1])
		outjoin.RightColumn = strings.TrimSpace(rightColumn[1])

		joiner = append(joiner, outjoin)
	}
	return joiner
}

func (where *DBWhere) parseNumber(segment string, i int) int {
	var buffer string
	for ; i < len(segment); i++ {
		char := segment[i]
		if '0' <= char && char <= '9' {
			buffer += string(char)
		} else {
			i--
			where.Expr = append(where.Expr, DBToken{buffer, "number"})
			return i
		}
	}
	return i
}

func (where *DBWhere) parseColumn(segment string, i int) int {
	var buffer string
	for ; i < len(segment); i++ {
		char := segment[i]
		switch {
		case ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '.' || char == '_':
			buffer += string(char)
		case char == '(':
			return where.parseFunction(segment, buffer, i)
		default:
			i--
			where.Expr = append(where.Expr, DBToken{buffer, "column"})
			return i
		}
	}
	return i
}

func (where *DBWhere) parseFunction(segment string, buffer string, i int) int {
	var preI = i
	i = skipFunctionCall(segment, i-1)
	buffer += segment[preI:i] + string(segment[i])
	where.Expr = append(where.Expr, DBToken{buffer, "function"})
	return i
}

func (where *DBWhere) parseString(segment string, i int) int {
	var buffer string
	i++
	for ; i < len(segment); i++ {
		char := segment[i]
		if char != '\'' {
			buffer += string(char)
		} else {
			where.Expr = append(where.Expr, DBToken{buffer, "string"})
			return i
		}
	}
	return i
}

func (where *DBWhere) parseOperator(segment string, i int) int {
	var buffer string
	for ; i < len(segment); i++ {
		char := segment[i]
		if isOpByte(char) {
			buffer += string(char)
		} else {
			i--
			where.Expr = append(where.Expr, DBToken{buffer, "operator"})
			return i
		}
	}
	return i
}

// TODO: Make this case insensitive
func normalizeAnd(in string) string {
	in = strings.Replace(in, " and ", " AND ", -1)
	return strings.Replace(in, " && ", " AND ", -1)
}
func normalizeOr(in string) string {
	in = strings.Replace(in, " or ", " OR ", -1)
	return strings.Replace(in, " || ", " OR ", -1)
}

// TODO: Write tests for this
func processWhere(wherestr string) (where []DBWhere) {
	if wherestr == "" {
		return where
	}
	wherestr = normalizeAnd(wherestr)
	wherestr = normalizeOr(wherestr)

	for _, segment := range strings.Split(wherestr, " AND ") {
		var tmpWhere = &DBWhere{[]DBToken{}}
		segment += ")"
		for i := 0; i < len(segment); i++ {
			char := segment[i]
			switch {
			case '0' <= char && char <= '9':
				i = tmpWhere.parseNumber(segment, i)
			// TODO: Sniff the third byte offset from char or it's non-existent to avoid matching uppercase strings which start with OR
			case char == 'O' && (i+1) < len(segment) && segment[i+1] == 'R':
				tmpWhere.Expr = append(tmpWhere.Expr, DBToken{"OR", "or"})
				i += 1
			case ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_':
				i = tmpWhere.parseColumn(segment, i)
			case char == '\'':
				i = tmpWhere.parseString(segment, i)
			case char == ')' && i < (len(segment)-1):
				tmpWhere.Expr = append(tmpWhere.Expr, DBToken{")", "operator"})
			case isOpByte(char):
				i = tmpWhere.parseOperator(segment, i)
			case char == '?':
				tmpWhere.Expr = append(tmpWhere.Expr, DBToken{"?", "substitute"})
			}
		}
		where = append(where, *tmpWhere)
	}
	return where
}

func (setter *DBSetter) parseNumber(segment string, i int) int {
	var buffer string
	for ; i < len(segment); i++ {
		char := segment[i]
		if '0' <= char && char <= '9' {
			buffer += string(char)
		} else {
			setter.Expr = append(setter.Expr, DBToken{buffer, "number"})
			return i
		}
	}
	return i
}

func (setter *DBSetter) parseColumn(segment string, i int) int {
	var buffer string
	for ; i < len(segment); i++ {
		char := segment[i]
		switch {
		case ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_':
			buffer += string(char)
		case char == '(':
			return setter.parseFunction(segment, buffer, i)
		default:
			i--
			setter.Expr = append(setter.Expr, DBToken{buffer, "column"})
			return i
		}
	}
	return i
}

func (setter *DBSetter) parseFunction(segment string, buffer string, i int) int {
	var preI = i
	i = skipFunctionCall(segment, i-1)
	buffer += segment[preI:i] + string(segment[i])
	setter.Expr = append(setter.Expr, DBToken{buffer, "function"})
	return i
}

func (setter *DBSetter) parseString(segment string, i int) int {
	var buffer string
	i++
	for ; i < len(segment); i++ {
		char := segment[i]
		if char != '\'' {
			buffer += string(char)
		} else {
			setter.Expr = append(setter.Expr, DBToken{buffer, "string"})
			return i
		}
	}
	return i
}

func (setter *DBSetter) parseOperator(segment string, i int) int {
	var buffer string
	for ; i < len(segment); i++ {
		char := segment[i]
		if isOpByte(char) {
			buffer += string(char)
		} else {
			i--
			setter.Expr = append(setter.Expr, DBToken{buffer, "operator"})
			return i
		}
	}
	return i
}

func processSet(setstr string) (setter []DBSetter) {
	if setstr == "" {
		return setter
	}

	// First pass, splitting the string by commas while ignoring the innards of functions
	var setset []string
	var buffer string
	var lastItem int
	setstr += ","
	for i := 0; i < len(setstr); i++ {
		if setstr[i] == '(' {
			i = skipFunctionCall(setstr, i-1)
			setset = append(setset, setstr[lastItem:i+1])
			buffer = ""
			lastItem = i + 2
		} else if setstr[i] == ',' && buffer != "" {
			setset = append(setset, buffer)
			buffer = ""
			lastItem = i + 1
		} else if (setstr[i] > 32) && setstr[i] != ',' && setstr[i] != ')' {
			buffer += string(setstr[i])
		}
	}

	// Second pass. Break this setitem into manageable chunks
	for _, setitem := range setset {
		halves := strings.Split(setitem, "=")
		if len(halves) != 2 {
			continue
		}
		tmpSetter := &DBSetter{Column: strings.TrimSpace(halves[0])}
		segment := halves[1] + ")"

		for i := 0; i < len(segment); i++ {
			char := segment[i]
			switch {
			case '0' <= char && char <= '9':
				i = tmpSetter.parseNumber(segment, i)
			case ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_':
				i = tmpSetter.parseColumn(segment, i)
			case char == '\'':
				i = tmpSetter.parseString(segment, i)
			case isOpByte(char):
				i = tmpSetter.parseOperator(segment, i)
			case char == '?':
				tmpSetter.Expr = append(tmpSetter.Expr, DBToken{"?", "substitute"})
			}
		}
		setter = append(setter, *tmpSetter)
	}
	return setter
}

func processLimit(limitstr string) (limiter DBLimit) {
	halves := strings.Split(limitstr, ",")
	if len(halves) == 2 {
		limiter.Offset = halves[0]
		limiter.MaxCount = halves[1]
	} else {
		limiter.MaxCount = halves[0]
	}
	return limiter
}

func isOpByte(char byte) bool {
	return char == '<' || char == '>' || char == '=' || char == '!' || char == '*' || char == '%' || char == '+' || char == '-' || char == '/' || char == '(' || char == ')'
}

func isOpRune(char rune) bool {
	return char == '<' || char == '>' || char == '=' || char == '!' || char == '*' || char == '%' || char == '+' || char == '-' || char == '/' || char == '(' || char == ')'
}

func processFields(fieldstr string) (fields []DBField) {
	if fieldstr == "" {
		return fields
	}
	var buffer string
	var lastItem int
	fieldstr += ","
	for i := 0; i < len(fieldstr); i++ {
		if fieldstr[i] == '(' {
			i = skipFunctionCall(fieldstr, i-1)
			fields = append(fields, DBField{Name: fieldstr[lastItem : i+1], Type: getIdentifierType(fieldstr[lastItem : i+1])})
			buffer = ""
			lastItem = i + 2
		} else if fieldstr[i] == ',' && buffer != "" {
			fields = append(fields, DBField{Name: buffer, Type: getIdentifierType(buffer)})
			buffer = ""
			lastItem = i + 1
		} else if (fieldstr[i] >= 32) && fieldstr[i] != ',' && fieldstr[i] != ')' {
			buffer += string(fieldstr[i])
		}
	}
	return fields
}

func getIdentifierType(identifier string) string {
	if ('a' <= identifier[0] && identifier[0] <= 'z') || ('A' <= identifier[0] && identifier[0] <= 'Z') {
		if identifier[len(identifier)-1] == ')' {
			return "function"
		}
		return "column"
	}
	if identifier[0] == '\'' || identifier[0] == '"' {
		return "string"
	}
	return "literal"
}

func getIdentifier(segment string, startOffset int) (out string, i int) {
	segment = strings.TrimSpace(segment)
	segment += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(segment); i++ {
		if segment[i] == '(' {
			i = skipFunctionCall(segment, i)
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
		if (segment[i] == ' ' || isOpByte(segment[i])) && i != startOffset {
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(segment[startOffset:]), (i - 1)
}

func getOperator(segment string, startOffset int) (out string, i int) {
	segment = strings.TrimSpace(segment)
	segment += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(segment); i++ {
		if !isOpByte(segment[i]) && i != startOffset {
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(segment[startOffset:]), (i - 1)
}

func skipFunctionCall(data string, index int) int {
	var braceCount int
	for ; index < len(data); index++ {
		char := data[index]
		if char == '(' {
			braceCount++
		} else if char == ')' {
			braceCount--
			if braceCount == 0 {
				return index
			}
		}
	}
	return index
}

func writeFile(name string, content string) (err error) {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	return f.Close()
}
