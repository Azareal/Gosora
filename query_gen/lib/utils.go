/* WIP Under Construction */
package qgen

//import "fmt"
import "strings"
import "os"

func _process_columns(colstr string) (columns []DB_Column) {
	if colstr == "" {
		return columns
	}
	colstr = strings.Replace(colstr," as "," AS ",-1)
	for _, segment := range strings.Split(colstr,",") {
		var outcol DB_Column
		dothalves := strings.Split(strings.TrimSpace(segment),".")
		
		var halves []string
		if len(dothalves) == 2 {
			outcol.Table = dothalves[0]
			halves = strings.Split(dothalves[1]," AS ")
		} else {
			halves = strings.Split(dothalves[0]," AS ")
		}
		
		halves[0] = strings.TrimSpace(halves[0])
		if len(halves) == 2 {
			outcol.Alias = strings.TrimSpace(halves[1])
		}
		if halves[0][len(halves[0]) - 1] == ')' {
			outcol.Type = "function"
		} else {
			outcol.Type = "column"
		}
		
		outcol.Left = halves[0]
		columns = append(columns,outcol)
	}
	return columns
}

func _process_orderby(orderstr string) (order []DB_Order) {
	if orderstr == "" {
		return order
	}
	for _, segment := range strings.Split(orderstr,",") {
		var outorder DB_Order
		halves := strings.Split(strings.TrimSpace(segment)," ")
		if len(halves) != 2 {
			continue
		}
		outorder.Column = halves[0]
		outorder.Order = strings.ToLower(halves[1])
		order = append(order,outorder)
	}
	return order
}

func _process_joiner(joinstr string) (joiner []DB_Joiner) {
	if joinstr == "" {
		return joiner
	}
	joinstr = strings.Replace(joinstr," on "," ON ",-1)
	joinstr = strings.Replace(joinstr," and "," AND ",-1)
	for _, segment := range strings.Split(joinstr," AND ") {
		var outjoin DB_Joiner
		var parseOffset int
		var left, right string
		
		left, parseOffset = _get_identifier(segment, parseOffset)
		outjoin.Operator, parseOffset = _get_operator(segment, parseOffset + 1)
		right, parseOffset = _get_identifier(segment, parseOffset + 1)
		
		
		left_column := strings.Split(left,".")
		right_column := strings.Split(right,".")
		outjoin.LeftTable = strings.TrimSpace(left_column[0])
		outjoin.RightTable = strings.TrimSpace(right_column[0])
		outjoin.LeftColumn = strings.TrimSpace(left_column[1])
		outjoin.RightColumn = strings.TrimSpace(right_column[1])
		
		joiner = append(joiner,outjoin)
	}
	return joiner
}

// TO-DO: Add support for keywords like BETWEEN. We'll probably need an arbitrary expression parser like with the update setters.
func _process_where(wherestr string) (where []DB_Where) {
	if wherestr == "" {
		return where
	}
	wherestr = strings.Replace(wherestr," and "," AND ",-1)
	for _, segment := range strings.Split(wherestr," AND ") {
		// TO-DO: Subparse the contents of a function and spit out a DB_Function struct
		var outwhere DB_Where
		var parseOffset int
		var left, right string
		
		left, parseOffset = _get_identifier(segment, parseOffset)
		outwhere.Operator, parseOffset = _get_operator(segment, parseOffset + 1)
		right, parseOffset = _get_identifier(segment, parseOffset + 1)
		outwhere.LeftType = _get_identifier_type(left)
		outwhere.RightType = _get_identifier_type(right)
		
		left_operand := strings.Split(left,".")
		right_operand := strings.Split(right,".")
		
		if len(left_operand) == 2 {
			outwhere.LeftTable = strings.TrimSpace(left_operand[0])
			outwhere.LeftColumn = strings.TrimSpace(left_operand[1])
		} else {
			outwhere.LeftColumn = strings.TrimSpace(left_operand[0])
		}
		
		if len(right_operand) == 2 {
			outwhere.RightTable = strings.TrimSpace(right_operand[0])
			outwhere.RightColumn = strings.TrimSpace(right_operand[1])
		} else {
			outwhere.RightColumn = strings.TrimSpace(right_operand[0])
		}
		
		where = append(where,outwhere)
	}
	return where
}

func _process_set(setstr string) (setter []DB_Setter) {
	if setstr == "" {
		return setter
	}
	//fmt.Println("setstr",setstr)
	
	// First pass, splitting the string by commas while ignoring the innards of functions
	var setset []string
	var buffer string
	var last_item int
	setstr += ","
	for i := 0; i < len(setstr); i++  {
		if setstr[i] == '(' {
			i = _skip_function_call(setstr,i-1)
			setset = append(setset,setstr[last_item:i+1])
			buffer = ""
			last_item = i + 2
		} else if setstr[i] == ',' && buffer != "" {
			setset = append(setset,buffer)
			buffer = ""
			last_item = i + 1
		} else if (setstr[i] > 32) && setstr[i] != ',' && setstr[i] != ')' {
			buffer += string(setstr[i])
		}
	}
	
	// Second pass. Break this setitem into manageable chunks
	buffer = ""
	for _, setitem := range setset {
		var tmp_setter DB_Setter
		halves := strings.Split(setitem,"=")
		if len(halves) != 2 {
			continue
		}
		tmp_setter.Column = strings.TrimSpace(halves[0])
		
		halves[1] += ")"
		var optype int // 0: None, 1: Number, 2: Column, 3: Function, 4: String, 5: Operator
		//fmt.Println("halves[1]",halves[1])
		for i := 0; i < len(halves[1]); i++ {
			char := halves[1][i]
			//fmt.Println("optype",optype)
			switch(optype) {
			case 0: // unknown
				if ('0' <= char && char <= '9') {
					optype = 1
					buffer = string(char)
				} else if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') {
					optype = 2
					buffer = string(char)
				} else if char == '\'' {
					optype = 4
				} else if _is_op_byte(char) {
					optype = 5
					buffer = string(char)
				} else if char == '?' {
					//fmt.Println("Expr:","?")
					tmp_setter.Expr = append(tmp_setter.Expr,DB_Token{"?","substitute"})
				}
			case 1: // number
				if ('0' <= char && char <= '9') {
					buffer += string(char)
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmp_setter.Expr = append(tmp_setter.Expr,DB_Token{buffer,"number"})
				}
			case 2: // column
				if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') {
					buffer += string(char)
				} else if char == '(' {
					optype = 3
					i--
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmp_setter.Expr = append(tmp_setter.Expr,DB_Token{buffer,"column"})
				}
			case 3: // function
				var pre_i int = i
				//fmt.Println("buffer",buffer)
				//fmt.Println("len(halves)",len(halves[1]))
				//fmt.Println("pre_i",string(halves[1][pre_i]))
				//fmt.Println("msg prior to pre_i",halves[1][0:pre_i])
				i = _skip_function_call(halves[1],i-1)
				//fmt.Println("i",i)
				//fmt.Println("msg prior to i-1",halves[1][0:i-1])
				//fmt.Println("string(i-1)",string(halves[1][i-1]))
				//fmt.Println("string(i)",string(halves[1][i]))
				buffer += halves[1][pre_i:i] + string(halves[1][i])
				//fmt.Println("Expr:",buffer)
				tmp_setter.Expr = append(tmp_setter.Expr,DB_Token{buffer,"function"})
				optype = 0
			case 4: // string
				if char != '\'' {
					buffer += string(char)
				} else {
					optype = 0
					//fmt.Println("Expr:",buffer)
					tmp_setter.Expr = append(tmp_setter.Expr,DB_Token{buffer,"string"})
				}
			case 5: // operator
				if _is_op_byte(char) {
					buffer += string(char)
				} else {
					optype = 0
					i--
					//fmt.Println("Expr:",buffer)
					tmp_setter.Expr = append(tmp_setter.Expr,DB_Token{buffer,"operator"})
				}
			}
		}
		setter = append(setter,tmp_setter)
	}
	//fmt.Println("setter",setter)
	return setter
}

func _process_limit(limitstr string) (limiter DB_Limit) {
	halves := strings.Split(limitstr,",")
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

func _process_fields(fieldstr string) (fields []DB_Field) {
	if fieldstr == "" {
		return fields
	}
	var buffer string
	var last_item int
	fieldstr += ","
	for i := 0; i < len(fieldstr); i++  {
		if fieldstr[i] == '(' {
			i = _skip_function_call(fieldstr,i-1)
			fields = append(fields,DB_Field{Name:fieldstr[last_item:i+1],Type:_get_identifier_type(fieldstr[last_item:i+1])})
			buffer = ""
			last_item = i + 2
		} else if fieldstr[i] == ',' && buffer != "" {
			fields = append(fields,DB_Field{Name:buffer,Type:_get_identifier_type(buffer)})
			buffer = ""
			last_item = i + 1
		} else if (fieldstr[i] > 32) && fieldstr[i] != ',' && fieldstr[i] != ')' {
			buffer += string(fieldstr[i])
		}
	}
	return fields
}

func _get_identifier_type(identifier string) string {
	if ('a' <= identifier[0] && identifier[0] <= 'z') || ('A' <= identifier[0] && identifier[0] <= 'Z') {
		if identifier[len(identifier) - 1] == ')' {
			return "function"
		}
		return "column"
	}
	if identifier[0] == '\'' || identifier[0] == '"' {
		return "string"
	}
	return "literal"
}

func _get_identifier(segment string, startOffset int) (out string, i int) {
	segment = strings.TrimSpace(segment)
	segment += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(segment); i++ {
		if segment[i] == '(' {
			i = _skip_function_call(segment,i)
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
		if (segment[i] == ' ' || _is_op_byte(segment[i])) && i != startOffset {
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(segment[startOffset:]), (i - 1)
}

func _get_operator(segment string, startOffset int) (out string, i int) {
	segment = strings.TrimSpace(segment)
	segment += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(segment); i++ {
		if !_is_op_byte(segment[i]) && i != startOffset {
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(segment[startOffset:]), (i - 1)
}

func _skip_function_call(data string, index int) int {
	var brace_count int
	for ;index < len(data); index++{
		char := data[index]
		if char == '(' {
			brace_count++
		} else if char == ')' {
			brace_count--
			if brace_count == 0 {
				return index
			}
		}
	}
	return index
}

func write_file(name string, content string) (err error) {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	f.Sync()
	f.Close()
	return
}
