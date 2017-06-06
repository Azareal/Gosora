/* WIP Under Construction */
package main

//import "fmt"
import "strings"
import "log"
import "os"

var db_registry []DB_Adapter
var blank_order []DB_Order

type DB_Column struct
{
	Table string
	Left string // Could be a function or a column, so I'm naming this Left
	Alias string // aka AS Blah, if it's present
	Type string // function or column
}

type DB_Where struct
{
	Left string
	Right string
	Operator string
	LeftType string
	RightType string
}

type DB_Joiner struct
{
	LeftTable string
	LeftColumn string
	RightTable string
	RightColumn string
}

type DB_Order struct
{
	Column string
	Order string
}

type DB_Adapter interface {
	get_name() string
	simple_insert(string,string,string,[]string,[]bool) error
	//simple_replace(string,string,[]string,[]string,[]bool) error
	simple_update() error
	simple_select(string,string,string,string,string/*,int,int*/) error
	simple_left_join(string,string,string,string,string,string,string/*,int,int*/) error
	write() error
	// TO-DO: Add a simple query builder
}

func main() {
	log.Println("Running the query generator")
	for _, adapter := range db_registry {
		log.Println("Building the queries for the " + adapter.get_name() + " adapter")
		write_statements(adapter)
		adapter.write()
	}
}

func write_statements(adapter DB_Adapter) error {
	// url_prefix and url_name will be removed from this query in a later commit
	adapter.simple_select("get_user","users","name, group, is_super_admin, avatar, message, url_prefix, url_name, level","uid = ?","")
	
	adapter.simple_select("get_full_user","users","name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip","uid = ?","")
		
	adapter.simple_select("get_topic","topics","title, content, createdBy, createdAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount","tid = ?","")
	
	adapter.simple_select("get_reply","replies","content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount","rid = ?","")
		
	adapter.simple_select("login","users","uid, name, password, salt","name = ?","")
		
	adapter.simple_select("get_password","users","password,salt","uid = ?","")
	
	adapter.simple_select("username_exists","users","name","name = ?","")
	
	adapter.simple_select("get_settings","settings","name, content, type","","")
	
	adapter.simple_select("get_setting","settings","content, type","name = ?","")
	
	adapter.simple_select("get_full_setting","settings","name, type, constraints","name = ?","")
	
	adapter.simple_select("is_plugin_active","plugins","active","uname = ?","")
	
	adapter.simple_left_join("get_topic_list","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	
	
	
	return nil
}

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
		halves := strings.Split(segment,"=")
		if len(halves) != 2 {
			continue
		}
		
		left_column := strings.Split(halves[0],".")
		right_column := strings.Split(halves[1],".")
		outjoin.LeftTable = strings.TrimSpace(left_column[0])
		outjoin.RightTable = strings.TrimSpace(right_column[0])
		outjoin.LeftColumn = strings.TrimSpace(left_column[1])
		outjoin.RightColumn = strings.TrimSpace(right_column[1])
		
		joiner = append(joiner,outjoin)
	}
	return joiner
}

func _process_where(wherestr string) (where []DB_Where) {
	if wherestr == "" {
		return where
	}
	wherestr = strings.Replace(wherestr," and "," AND ",-1)
	for _, segment := range strings.Split(wherestr," AND ") {
		// TO-DO: Subparse the contents of a function and spit out a DB_Function struct
		var outwhere DB_Where
		var parseOffset int
		outwhere.Left, parseOffset = _get_identifier(segment, parseOffset)
		outwhere.Operator, parseOffset = _get_operator(segment, parseOffset + 1)
		outwhere.Right, parseOffset = _get_identifier(segment, parseOffset + 1)
		outwhere.LeftType = _get_identifier_type(outwhere.Left)
		outwhere.RightType = _get_identifier_type(outwhere.Right)
		where = append(where,outwhere)
	}
	return where
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
	//fmt.Println("entering _get_identifier")
	segment = strings.TrimSpace(segment)
	segment += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(segment); i++ {
		if segment[i] == '(' {
			i = _skip_function_call(segment,i)
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
		if segment[i] == ' ' && i != startOffset {
			//fmt.Println("segment[startOffset:i]",segment[startOffset:i])
			//fmt.Println("startOffset",startOffset)
			//fmt.Println("segment[startOffset]",string(segment[startOffset]))
			//fmt.Println("i",i)
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(segment[startOffset:]), (i - 1)
}

func _get_operator(segment string, startOffset int) (out string, i int) {
	segment = strings.TrimSpace(segment)
	segment += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(segment); i++ {
		if segment[i] == ' ' && i != startOffset {
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
	}
	return strings.TrimSpace(segment[startOffset:]), (i - 1)
}

func _skip_function_call(segment string, i int) int {
	var brace_count int = 1
	for ; i < len(segment); i++ {
		if segment[i] == '(' {
			brace_count++
		} else if segment[i] == ')' {
			brace_count--
		}
		if brace_count == 0 {
			return i
		}
	}
	return i
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
