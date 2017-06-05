/* WIP Under Construction */
package main

import "fmt"
import "strings"
import "log"
import "os"

var db_registry []DB_Adapter
var blank_order []DB_Order

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
	Left string
	Right string
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
	simple_select(string,string,string,string,[]DB_Order/*,int,int*/) error
	simple_left_join(string,string,string,string,[]DB_Joiner,string,[]DB_Order/*,int,int*/) error
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
	adapter.simple_select("get_user","users","name, group, is_super_admin, avatar, message, url_prefix, url_name, level","uid = ?",blank_order)
	
	adapter.simple_select("get_full_user","users","name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip","uid = ?",blank_order)
		
	adapter.simple_select("get_topic","topics","title, content, createdBy, createdAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount","tid = ?",blank_order)
	
	adapter.simple_select("get_reply","replies","content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount","rid = ?",blank_order)
		
	adapter.simple_select("login","users","uid, name, password, salt","name = ?",blank_order)
		
	adapter.simple_select("get_password","users","password,salt","uid = ?",blank_order)
	
	adapter.simple_select("username_exists","users","name","name = ?",blank_order)
	
	/*
get_topic_list_stmt, err = db.Prepare("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
// A visual reference for me to glance at while I design this thing
*/
	//func (adapter *Mysql_Adapter) simple_left_join(name string, table1 string, table2 string, columns string, joiners []DB_Joiner, where []DB_Where, orderby []DB_Order/*, offset int, maxCount int*/) error {
	
	/*adapter.simple_left_join("get_topic_list","topics","users",
		"topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar",
		[]DB_Joiner{}
	)*/
	
	return nil
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
	fmt.Println("entering _get_identifier")
	segment = strings.TrimSpace(segment)
	segment += " " // Avoid overflow bugs with slicing
	for i = startOffset; i < len(segment); i++ {
		if segment[i] == '(' {
			i = _skip_function_call(segment,i)
			return strings.TrimSpace(segment[startOffset:i]), (i - 1)
		}
		if segment[i] == ' ' && i != startOffset {
			fmt.Println("segment[startOffset:i]",segment[startOffset:i])
			fmt.Println("startOffset",startOffset)
			fmt.Println("segment[startOffset]",string(segment[startOffset]))
			fmt.Println("i",i)
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
