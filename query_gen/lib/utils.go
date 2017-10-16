/*
*
*	Query Generator Library
*	WIP Under Construction
*	Copyright Azareal 2017 - 2018
*
 */
package qgen

//import "fmt"
import (
	"os"
	"strings"
)

func processColumns(colstr string) (columns []DB_Column) {
	if colstr == "" {
		return columns
	}
	colstr = strings.Replace(colstr, " as ", " AS ", -1)
	for _, segment := range strings.Split(colstr, ",") {
		var outcol DB_Column
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
func processOrderby(orderstr string) (order []DB_Order) {
	if orderstr == "" {
		return order
	}
	for _, segment := range strings.Split(orderstr, ",") {
		var outorder DB_Order
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

func processJoiner(joinstr string) (joiner []DB_Joiner) {
	if joinstr == "" {
		return joiner
	}
	joinstr = strings.Replace(joinstr, " on ", " ON ", -1)
	joinstr = strings.Replace(joinstr, " and ", " AND ", -1)
	for _, segment := range strings.Split(joinstr, " AND ") {
		var outjoin DB_Joiner
		var parseOffset int
		var left, right string

		left, parseOffset = getIdentifier(segment, parseOffset)
		outjoin.Operator, parseOffset = getOperator(segment, parseOffset+1)
		right, parseOffset = getIdentifier(segment, parseOffset+1)

		left_column := strings.Split(left, ".")
		right_column := strings.Split(right, ".")
		outjoin.LeftTable = strings.TrimSpace(left_column[0])
		outjoin.RightTable = strings.TrimSpace(right_column[0])
		outjoin.LeftColumn = strings.TrimSpace(left_column[1])
		outjoin.RightColumn = strings.TrimSpace(right_column[1])

		joiner = append(joiner, outjoin)
	}
	return joiner
}

func processWhere(wherestr string) (where []DB_Where) {
	if wherestr == "" {
		return where
	}
	wherestr = strings.Replace(wherestr, " and ", " AND ", -1)

	var buffer string
	var optype int // 0: None, 1: Number, 2: Column, 3: Function, 4: String, 5: Operator
	for _, segment := range strings.Split(wherestr, " AND ") {
		var tmp_where DB_Where
		segment += ")"
		for i := 0; i < len(segment); i++ {
			char := segment[i]
			//fmt.Println("optype",optype)
			switch optype {
			case 0: // unknown
				//fmt.Println("case 0:",char,string(char))
				if '0' <= char && char <= '9' {
					optype = 1
					buffer = string(char)
				} else if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_' {
					optype = 2
					buffer = string(char)
				} else if char == '\'' {
					optype = 4
					buffer = ""
				} else if _is_op_byte(char) {
					optype = 5
					buffer = string(char)
				} else if char == '?' {
					//fmt.Println("Expr:","?")
					tmp_where.Expr = append(tmp_where.Expr, DB_Token{"?", "substitute"})
				}
			case 1: // number
				if '0' <= char && char <= '9' {
					buffer += string(char)
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmp_where.Expr = append(tmp_where.Expr, DB_Token{buffer, "number"})
				}
			case 2: // column
				if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '.' || char == '_' {
					buffer += string(char)
				} else if char == '(' {
					optype = 3
					i--
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmp_where.Expr = append(tmp_where.Expr, DB_Token{buffer, "column"})
				}
			case 3: // function
				var preI = i
				//fmt.Println("buffer",buffer)
				//fmt.Println("len(halves)",len(halves[1]))
				//fmt.Println("preI",string(halves[1][preI]))
				//fmt.Println("msg prior to preI",halves[1][0:preI])
				i = skipFunctionCall(segment, i-1)
				//fmt.Println("i",i)
				//fmt.Println("msg prior to i-1",halves[1][0:i-1])
				//fmt.Println("string(i-1)",string(halves[1][i-1]))
				//fmt.Println("string(i)",string(halves[1][i]))
				buffer += segment[preI:i] + string(segment[i])
				//fmt.Println("Expr:",buffer)
				tmp_where.Expr = append(tmp_where.Expr, DB_Token{buffer, "function"})
				optype = 0
			case 4: // string
				if char != '\'' {
					buffer += string(char)
				} else {
					optype = 0
					//fmt.Println("Expr:",buffer)
					tmp_where.Expr = append(tmp_where.Expr, DB_Token{buffer, "string"})
				}
			case 5: // operator
				if _is_op_byte(char) {
					buffer += string(char)
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmp_where.Expr = append(tmp_where.Expr, DB_Token{buffer, "operator"})
				}
			default:
				panic("Bad optype in _process_where")
			}
		}
		where = append(where, tmp_where)
	}
	return where
}

func processSet(setstr string) (setter []DB_Setter) {
	if setstr == "" {
		return setter
	}
	//fmt.Println("setstr",setstr)

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
	buffer = ""
	for _, setitem := range setset {
		var tmpSetter DB_Setter
		halves := strings.Split(setitem, "=")
		if len(halves) != 2 {
			continue
		}
		tmpSetter.Column = strings.TrimSpace(halves[0])

		halves[1] += ")"
		var optype int // 0: None, 1: Number, 2: Column, 3: Function, 4: String, 5: Operator
		//fmt.Println("halves[1]",halves[1])
		for i := 0; i < len(halves[1]); i++ {
			char := halves[1][i]
			//fmt.Println("optype",optype)
			switch optype {
			case 0: // unknown
				if '0' <= char && char <= '9' {
					optype = 1
					buffer = string(char)
				} else if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_' {
					optype = 2
					buffer = string(char)
				} else if char == '\'' {
					optype = 4
					buffer = ""
				} else if _is_op_byte(char) {
					optype = 5
					buffer = string(char)
				} else if char == '?' {
					//fmt.Println("Expr:","?")
					tmpSetter.Expr = append(tmpSetter.Expr, DB_Token{"?", "substitute"})
				}
			case 1: // number
				if '0' <= char && char <= '9' {
					buffer += string(char)
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmpSetter.Expr = append(tmpSetter.Expr, DB_Token{buffer, "number"})
				}
			case 2: // column
				if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_' {
					buffer += string(char)
				} else if char == '(' {
					optype = 3
					i--
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmpSetter.Expr = append(tmpSetter.Expr, DB_Token{buffer, "column"})
				}
			case 3: // function
				var preI = i
				//fmt.Println("buffer",buffer)
				//fmt.Println("len(halves)",len(halves[1]))
				//fmt.Println("preI",string(halves[1][preI]))
				//fmt.Println("msg prior to preI",halves[1][0:preI])
				i = skipFunctionCall(halves[1], i-1)
				//fmt.Println("i",i)
				//fmt.Println("msg prior to i-1",halves[1][0:i-1])
				//fmt.Println("string(i-1)",string(halves[1][i-1]))
				//fmt.Println("string(i)",string(halves[1][i]))
				buffer += halves[1][preI:i] + string(halves[1][i])
				//fmt.Println("Expr:",buffer)
				tmpSetter.Expr = append(tmpSetter.Expr, DB_Token{buffer, "function"})
				optype = 0
			case 4: // string
				if char != '\'' {
					buffer += string(char)
				} else {
					optype = 0
					//fmt.Println("Expr:",buffer)
					tmpSetter.Expr = append(tmpSetter.Expr, DB_Token{buffer, "string"})
				}
			case 5: // operator
				if _is_op_byte(char) {
					buffer += string(char)
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmpSetter.Expr = append(tmpSetter.Expr, DB_Token{buffer, "operator"})
				}
			default:
				panic("Bad optype in _process_set")
			}
		}
		setter = append(setter, tmpSetter)
	}
	//fmt.Println("setter",setter)
	return setter
}

func processLimit(limitstr string) (limiter DB_Limit) {
	halves := strings.Split(limitstr, ",")
	if len(halves) == 2 {
		limiter.Offset = halves[0]
		limiter.MaxCount = halves[1]
	} else {
		limiter.MaxCount = halves[0]
	}
	return limiter
}

func _is_op_byte(char byte) bool {
	return char == '<' || char == '>' || char == '=' || char == '!' || char == '*' || char == '%' || char == '+' || char == '-' || char == '/'
}

func _is_op_rune(char rune) bool {
	return char == '<' || char == '>' || char == '=' || char == '!' || char == '*' || char == '%' || char == '+' || char == '-' || char == '/'
}

func processFields(fieldstr string) (fields []DB_Field) {
	if fieldstr == "" {
		return fields
	}
	var buffer string
	var lastItem int
	fieldstr += ","
	for i := 0; i < len(fieldstr); i++ {
		if fieldstr[i] == '(' {
			i = skipFunctionCall(fieldstr, i-1)
			fields = append(fields, DB_Field{Name: fieldstr[lastItem : i+1], Type: getIdentifierType(fieldstr[lastItem : i+1])})
			buffer = ""
			lastItem = i + 2
		} else if fieldstr[i] == ',' && buffer != "" {
			fields = append(fields, DB_Field{Name: buffer, Type: getIdentifierType(buffer)})
			buffer = ""
			lastItem = i + 1
		} else if (fieldstr[i] > 32) && fieldstr[i] != ',' && fieldstr[i] != ')' {
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
		if (segment[i] == ' ' || _is_op_byte(segment[i])) && i != startOffset {
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(segment[startOffset:]), (i - 1)
}

func getOperator(segment string, startOffset int) (out string, i int) {
	segment = strings.TrimSpace(segment)
	segment += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(segment); i++ {
		if !_is_op_byte(segment[i]) && i != startOffset {
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
