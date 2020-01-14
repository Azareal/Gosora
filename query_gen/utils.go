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
func processColumns(colStr string) (columns []DBColumn) {
	if colStr == "" {
		return columns
	}
	colStr = strings.Replace(colStr, " as ", " AS ", -1)
	for _, segment := range strings.Split(colStr, ",") {
		var outCol DBColumn
		dotHalves := strings.Split(strings.TrimSpace(segment), ".")

		var halves []string
		if len(dotHalves) == 2 {
			outCol.Table = dotHalves[0]
			halves = strings.Split(dotHalves[1], " AS ")
		} else {
			halves = strings.Split(dotHalves[0], " AS ")
		}

		halves[0] = strings.TrimSpace(halves[0])
		if len(halves) == 2 {
			outCol.Alias = strings.TrimSpace(halves[1])
		}
		if halves[0][len(halves[0])-1] == ')' {
			outCol.Type = TokenFunc
		} else if halves[0] == "?" {
			outCol.Type = TokenSub
		} else {
			outCol.Type = TokenColumn
		}

		outCol.Left = halves[0]
		columns = append(columns, outCol)
	}
	return columns
}

// TODO: Allow order by statements without a direction
func processOrderby(orderStr string) (order []DBOrder) {
	if orderStr == "" {
		return order
	}
	for _, segment := range strings.Split(orderStr, ",") {
		var outOrder DBOrder
		halves := strings.Split(strings.TrimSpace(segment), " ")
		if len(halves) != 2 {
			continue
		}
		outOrder.Column = halves[0]
		outOrder.Order = strings.ToLower(halves[1])
		order = append(order, outOrder)
	}
	return order
}

func processJoiner(joinStr string) (joiner []DBJoiner) {
	if joinStr == "" {
		return joiner
	}
	joinStr = strings.Replace(joinStr, " on ", " ON ", -1)
	joinStr = strings.Replace(joinStr, " and ", " AND ", -1)
	for _, segment := range strings.Split(joinStr, " AND ") {
		var outJoin DBJoiner
		var parseOffset int
		var left, right string

		left, parseOffset = getIdentifier(segment, parseOffset)
		outJoin.Operator, parseOffset = getOperator(segment, parseOffset+1)
		right, parseOffset = getIdentifier(segment, parseOffset+1)

		leftColumn := strings.Split(left, ".")
		rightColumn := strings.Split(right, ".")
		outJoin.LeftTable = strings.TrimSpace(leftColumn[0])
		outJoin.RightTable = strings.TrimSpace(rightColumn[0])
		outJoin.LeftColumn = strings.TrimSpace(leftColumn[1])
		outJoin.RightColumn = strings.TrimSpace(rightColumn[1])

		joiner = append(joiner, outJoin)
	}
	return joiner
}

func (wh *DBWhere) parseNumber(seg string, i int) int {
	//var buffer string
	si := i
	l := 0
	for ; i < len(seg); i++ {
		ch := seg[i]
		if '0' <= ch && ch <= '9' {
			//buffer += string(ch)
			l++
		} else {
			i--
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			wh.Expr = append(wh.Expr, DBToken{str, TokenNumber})
			return i
		}
	}
	return i
}

func (wh *DBWhere) parseColumn(seg string, i int) int {
	//var buffer string
	si := i
	l := 0
	for ; i < len(seg); i++ {
		ch := seg[i]
		switch {
		case ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '.' || ch == '_':
			//buffer += string(ch)
			l++
		case ch == '(':
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			return wh.parseFunction(seg, str, i)
		default:
			i--
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			wh.Expr = append(wh.Expr, DBToken{str, TokenColumn})
			return i
		}
	}
	return i
}

func (wh *DBWhere) parseFunction(seg string, buffer string, i int) int {
	preI := i
	i = skipFunctionCall(seg, i-1)
	buffer += seg[preI:i] + string(seg[i])
	wh.Expr = append(wh.Expr, DBToken{buffer, TokenFunc})
	return i
}

func (wh *DBWhere) parseString(seg string, i int) int {
	//var buffer string
	i++
	si := i
	l := 0
	for ; i < len(seg); i++ {
		ch := seg[i]
		if ch != '\'' {
			//buffer += string(ch)
			l++
		} else {
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			wh.Expr = append(wh.Expr, DBToken{str, TokenString})
			return i
		}
	}
	return i
}

func (wh *DBWhere) parseOperator(seg string, i int) int {
	//var buffer string
	si := i
	l := 0
	for ; i < len(seg); i++ {
		ch := seg[i]
		if isOpByte(ch) {
			//buffer += string(ch)
			l++
		} else {
			i--
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			wh.Expr = append(wh.Expr, DBToken{str, TokenOp})
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
func processWhere(whereStr string) (where []DBWhere) {
	if whereStr == "" {
		return where
	}
	whereStr = normalizeAnd(whereStr)
	whereStr = normalizeOr(whereStr)

	for _, seg := range strings.Split(whereStr, " AND ") {
		tmpWhere := &DBWhere{[]DBToken{}}
		seg += ")"
		for i := 0; i < len(seg); i++ {
			ch := seg[i]
			switch {
			case '0' <= ch && ch <= '9':
				i = tmpWhere.parseNumber(seg, i)
			// TODO: Sniff the third byte offset from char or it's non-existent to avoid matching uppercase strings which start with OR
			case ch == 'O' && (i+1) < len(seg) && seg[i+1] == 'R':
				tmpWhere.Expr = append(tmpWhere.Expr, DBToken{"OR", TokenOr})
				i += 1
			case ch == 'N' && (i+2) < len(seg) && seg[i+1] == 'O' && seg[i+2] == 'T':
				tmpWhere.Expr = append(tmpWhere.Expr, DBToken{"NOT", TokenNot})
				i += 2
			case ch == 'L' && (i+3) < len(seg) && seg[i+1] == 'I' && seg[i+2] == 'K' && seg[i+3] == 'E':
				tmpWhere.Expr = append(tmpWhere.Expr, DBToken{"LIKE", TokenLike})
				i += 3
			case ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_':
				i = tmpWhere.parseColumn(seg, i)
			case ch == '\'':
				i = tmpWhere.parseString(seg, i)
			case ch == ')' && i < (len(seg)-1):
				tmpWhere.Expr = append(tmpWhere.Expr, DBToken{")", TokenOp})
			case isOpByte(ch):
				i = tmpWhere.parseOperator(seg, i)
			case ch == '?':
				tmpWhere.Expr = append(tmpWhere.Expr, DBToken{"?", TokenSub})
			}
		}
		where = append(where, *tmpWhere)
	}
	return where
}

func (set *DBSetter) parseNumber(seg string, i int) int {
	//var buffer string
	si := i
	l := 0
	for ; i < len(seg); i++ {
		ch := seg[i]
		if '0' <= ch && ch <= '9' {
			//buffer += string(ch)
			l++
		} else {
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			set.Expr = append(set.Expr, DBToken{str, TokenNumber})
			return i
		}
	}
	return i
}

func (set *DBSetter) parseColumn(seg string, i int) int {
	//var buffer string
	si := i
	l := 0
	for ; i < len(seg); i++ {
		ch := seg[i]
		switch {
		case ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_':
			//buffer += string(ch)
			l++
		case ch == '(':
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			return set.parseFunction(seg, str, i)
		default:
			i--
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			set.Expr = append(set.Expr, DBToken{str, TokenColumn})
			return i
		}
	}
	return i
}

func (set *DBSetter) parseFunction(segment string, buffer string, i int) int {
	preI := i
	i = skipFunctionCall(segment, i-1)
	buffer += segment[preI:i] + string(segment[i])
	set.Expr = append(set.Expr, DBToken{buffer, TokenFunc})
	return i
}

func (set *DBSetter) parseString(seg string, i int) int {
	//var buffer string
	i++
	si := i
	l := 0
	for ; i < len(seg); i++ {
		ch := seg[i]
		if ch != '\'' {
			//buffer += string(ch)
			l++
		} else {
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			set.Expr = append(set.Expr, DBToken{str, TokenString})
			return i
		}
	}
	return i
}

func (set *DBSetter) parseOperator(seg string, i int) int {
	//var buffer string
	si := i
	l := 0
	for ; i < len(seg); i++ {
		ch := seg[i]
		if isOpByte(ch) {
			//buffer += string(ch)
			l++
		} else {
			i--
			var str string
			if l != 0 {
				str = seg[si : si+l]
			}
			set.Expr = append(set.Expr, DBToken{str, TokenOp})
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
			ch := segment[i]
			switch {
			case '0' <= ch && ch <= '9':
				i = tmpSetter.parseNumber(segment, i)
			case ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_':
				i = tmpSetter.parseColumn(segment, i)
			case ch == '\'':
				i = tmpSetter.parseString(segment, i)
			case isOpByte(ch):
				i = tmpSetter.parseOperator(segment, i)
			case ch == '?':
				tmpSetter.Expr = append(tmpSetter.Expr, DBToken{"?", TokenSub})
			}
		}
		setter = append(setter, *tmpSetter)
	}
	return setter
}

func processLimit(limitStr string) (limit DBLimit) {
	halves := strings.Split(limitStr, ",")
	if len(halves) == 2 {
		limit.Offset = halves[0]
		limit.MaxCount = halves[1]
	} else {
		limit.MaxCount = halves[0]
	}
	return limit
}

func isOpByte(ch byte) bool {
	return ch == '<' || ch == '>' || ch == '=' || ch == '!' || ch == '*' || ch == '%' || ch == '+' || ch == '-' || ch == '/' || ch == '(' || ch == ')'
}

func isOpRune(ch rune) bool {
	return ch == '<' || ch == '>' || ch == '=' || ch == '!' || ch == '*' || ch == '%' || ch == '+' || ch == '-' || ch == '/' || ch == '(' || ch == ')'
}

func processFields(fieldStr string) (fields []DBField) {
	if fieldStr == "" {
		return fields
	}
	var buffer string
	var lastItem int
	fieldStr += ","
	for i := 0; i < len(fieldStr); i++ {
		ch := fieldStr[i]
		if ch == '(' {
			i = skipFunctionCall(fieldStr, i-1)
			fields = append(fields, DBField{Name: fieldStr[lastItem : i+1], Type: getIdentifierType(fieldStr[lastItem : i+1])})
			buffer = ""
			lastItem = i + 2
		} else if ch == ',' && buffer != "" {
			fields = append(fields, DBField{Name: buffer, Type: getIdentifierType(buffer)})
			buffer = ""
			lastItem = i + 1
		} else if (ch >= 32) && ch != ',' && ch != ')' {
			buffer += string(ch)
		}
	}
	return fields
}

func getIdentifierType(iden string) int {
	if ('a' <= iden[0] && iden[0] <= 'z') || ('A' <= iden[0] && iden[0] <= 'Z') {
		if iden[len(iden)-1] == ')' {
			return IdenFunc
		}
		return IdenColumn
	}
	if iden[0] == '\'' || iden[0] == '"' {
		return IdenString
	}
	return IdenLiteral
}

func getIdentifier(seg string, startOffset int) (out string, i int) {
	seg = strings.TrimSpace(seg)
	seg += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(seg); i++ {
		ch := seg[i]
		if ch == '(' {
			i = skipFunctionCall(seg, i)
			return strings.TrimSpace(seg[startOffset:i]), (i - 1)
		}
		if (ch == ' ' || isOpByte(ch)) && i != startOffset {
			return strings.TrimSpace(seg[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(seg[startOffset:]), (i - 1)
}

func getOperator(seg string, startOffset int) (out string, i int) {
	seg = strings.TrimSpace(seg)
	seg += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(seg); i++ {
		if !isOpByte(seg[i]) && i != startOffset {
			return strings.TrimSpace(seg[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(seg[startOffset:]), (i - 1)
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

func writeFile(name, content string) (err error) {
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
