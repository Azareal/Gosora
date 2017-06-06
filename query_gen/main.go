/* WIP Under Construction */
package main

import "fmt"
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

type DB_Field struct
{
	Name string
	Type string
}

type DB_Where struct
{
	LeftTable string
	LeftColumn string
	RightTable string
	RightColumn string
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
	simple_insert(string,string,string,string) error
	//simple_replace(string,string,[]string,string) error
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
	
	adapter.simple_select("get_users","users","uid, name, group, active, is_super_admin, avatar","","")
	
	adapter.simple_select("is_theme_default","themes","default","uname = ?","")
	
	adapter.simple_select("get_modlogs","moderation_logs","action, elementID, elementType, ipaddress, actorID, doneAt","","")
	
	adapter.simple_select("get_reply_tid","replies","tid","rid = ?","")
	
	adapter.simple_select("get_topic_fid","topics","parentID","tid = ?","")
	
	adapter.simple_select("get_user_reply_uid","users_replies","uid","rid = ?","")
	
	adapter.simple_select("has_liked_topic","likes","targetItem","sentBy = ? and targetItem = ? and targetType = 'topics'","")
	
	/*"select targetItem from likes where sentBy = ? and targetItem = ? and targetType = 'replies'"*/
	
	
	
	adapter.simple_left_join("get_topic_list","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	
	adapter.simple_left_join("get_topic_user","topics","users","topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level","topics.createdBy = users.uid","tid = ?","")
	
	adapter.simple_left_join("get_topic_by_reply","replies","topics","topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount","replies.tid = topics.tid","rid = ?","")
	
	adapter.simple_left_join("get_topic_replies","replies","users","replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress","replies.createdBy = users.uid","tid = ?","")
	
	adapter.simple_left_join("get_forum_topics","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","topics.parentID = ?","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc")
	
	adapter.simple_left_join("get_profile_replies","users_replies","users","users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group","users_replies.createdBy = users.uid","users_replies.uid = ?","")
	
	adapter.simple_insert("create_topic","topics","parentID,title,content,parsed_content,createdAt,lastReplyAt,ipaddress,words,createdBy","?,?,?,?,NOW(),NOW(),?,?,?")
	
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

func _process_fields(fieldstr string) (fields []DB_Field) {
	fmt.Println("_Entering _process_fields")
	if fieldstr == "" {
		return fields
	}
	var buffer string
	var last_item int
	fieldstr += ","
	for i := 0; i < len(fieldstr); i++  {
		if fieldstr[i] == '(' {
			var pre_i int
			pre_i = i
			i = _skip_function_call(fieldstr,i-1)
			fmt.Println("msg prior to i",fieldstr[0:i])
			fmt.Println("len(fieldstr)",len(fieldstr))
			fmt.Println("pre_i",pre_i)
			fmt.Println("last_item",last_item)
			fmt.Println("pre_i",string(fieldstr[pre_i]))
			fmt.Println("last_item",string(fieldstr[last_item]))
			fmt.Println("fieldstr[pre_i:i+1]",fieldstr[pre_i:i+1])
			fmt.Println("fieldstr[last_item:i+1]",fieldstr[last_item:i+1])
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
	fmt.Println("fields",fields)
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

func _skip_function_call(data string, index int) int {
	var brace_count int
	for ;index < len(data); index++{
		char := data[index]
		if char == '(' {
			fmt.Println("Enter brace")
			brace_count++
		} else if char == ')' {
			brace_count--
			fmt.Println("Exit brace")
			if brace_count == 0 {
				fmt.Println("Exit function segment")
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
