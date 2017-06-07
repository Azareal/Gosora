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

type DB_Token struct {
	Contents string
	Type string // function, operator, column, number, string, substitute
}

type DB_Setter struct {
	Column string
	Expr []DB_Token // Simple expressions, the innards of functions are opaque for now.
}

type DB_Adapter interface {
	get_name() string
	simple_insert(string,string,string,string) error
	//simple_replace(string,string,[]string,string) error
	simple_update(string,string,string,string) error
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
	
	adapter.simple_select("has_liked_reply","likes","targetItem","sentBy = ? and targetItem = ? and targetType = 'replies'","")
	
	adapter.simple_select("get_user_name","users","name","uid = ?","")
	
	adapter.simple_select("get_emails_by_user","emails","email, validated","uid = ?","")
	
	adapter.simple_select("get_topic_basic","topics","title, content","tid = ?","")
	
	adapter.simple_left_join("get_topic_list","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	
	adapter.simple_left_join("get_topic_user","topics","users","topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level","topics.createdBy = users.uid","tid = ?","")
	
	adapter.simple_left_join("get_topic_by_reply","replies","topics","topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount","replies.tid = topics.tid","rid = ?","")
	
	adapter.simple_left_join("get_topic_replies","replies","users","replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress","replies.createdBy = users.uid","tid = ?","")
	
	adapter.simple_left_join("get_forum_topics","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","topics.parentID = ?","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc")
	
	adapter.simple_left_join("get_profile_replies","users_replies","users","users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group","users_replies.createdBy = users.uid","users_replies.uid = ?","")
	
	adapter.simple_insert("create_topic","topics","parentID,title,content,parsed_content,createdAt,lastReplyAt,ipaddress,words,createdBy","?,?,?,?,NOW(),NOW(),?,?,?")
	
	adapter.simple_insert("create_report","topics","title,content,parsed_content,createdAt,lastReplyAt,createdBy,data,parentID,css_class","?,?,?,NOW(),NOW(),?,?,1,'report'")

	adapter.simple_insert("create_reply","replies","tid,content,parsed_content,createdAt,ipaddress,words,createdBy","?,?,?,NOW(),?,?,?")
	
	adapter.simple_insert("create_action_reply","replies","tid,actionType,ipaddress,createdBy","?,?,?,?")
	
	adapter.simple_insert("create_like","likes","weight, targetItem, targetType, sentBy","?,?,?,?")
	
	adapter.simple_insert("add_activity","activity_stream","actor,targetUser,event,elementType,elementID","?,?,?,?,?")
	
	// Add an admin version of register_stmt with more flexibility?
	// create_account_stmt, err = db.Prepare("INSERT INTO
	adapter.simple_insert("register","users","name, email, password, salt, group, is_super_admin, session, active, message","?,?,?,?,?,0,?,?,''")
	
	
	adapter.simple_update("add_replies_to_topic","topics","postCount = postCount + ?, lastReplyAt = NOW()","tid = ?")
	
	adapter.simple_update("remove_replies_from_topic","topics","postCount = postCount - ?","tid = ?")
	
	adapter.simple_update("add_topics_to_forum","forums","topicCount = topicCount + ?","fid = ?")
	
	adapter.simple_update("remove_topics_from_forum","forums","topicCount = topicCount - ?","fid = ?")
	
	adapter.simple_update("update_forum_cache","forums","lastTopic = ?, lastTopicID = ?, lastReplyer = ?, lastReplyerID = ?, lastTopicTime = NOW()","fid = ?")

	adapter.simple_update("add_likes_to_topic","topics","likeCount = likeCount + ?","tid = ?")
	
	adapter.simple_update("add_likes_to_reply","replies","likeCount = likeCount + ?","rid = ?")
	
	adapter.simple_update("edit_topic","topics","title = ?, content = ?, parsed_content = ?, is_closed = ?","tid = ?")
	
	adapter.simple_update("edit_reply","replies","content = ?, parsed_content = ?","rid = ?")
	
	adapter.simple_update("stick_topic","topics","sticky = 1","tid = ?")
	
	adapter.simple_update("unstick_topic","topics","sticky = 0","tid = ?")
	
	adapter.simple_update("update_last_ip","users","last_ip = ?","uid = ?")

	adapter.simple_update("update_session","users","session = ?","uid = ?")
	
	adapter.simple_update("logout","users","session = ''","uid = ?")

	adapter.simple_update("set_password","users","password = ?, salt = ?","uid = ?")
	
	adapter.simple_update("set_avatar","users","avatar = ?","uid = ?")
	
	adapter.simple_update("set_username","users","name = ?","uid = ?")
	
	adapter.simple_update("change_group","users","group = ?","uid = ?")
	
	adapter.simple_update("activate_user","users","active = 1","uid = ?")
	
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
		if segment[i] == ' ' && i != startOffset {
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
