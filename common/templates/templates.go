package tmpl

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template/parse"
)

// TODO: Turn this file into a library
var textOverlapList = make(map[string]int)

// TODO: Stop hard-coding this here
var langPkg = "github.com/Azareal/Gosora/common/phrases"

type VarItem struct {
	Name        string
	Destination string
	Type        string
}

type VarItemReflect struct {
	Name        string
	Destination string
	Value       reflect.Value
}

type CTemplateConfig struct {
	Minify         bool
	Debug          bool
	SuperDebug     bool
	SkipHandles    bool
	SkipTmplPtrMap bool
	SkipInitBlock  bool
	PackageName    string
}

type OutBufferFrame struct {
	Body         string
	Type         string
	TemplateName string
}

type CContext struct {
	VarHolder    string
	HoldReflect  reflect.Value
	TemplateName string
	OutBuf       *[]OutBufferFrame
}

func (con *CContext) Push(nType string, body string) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, nType, con.TemplateName})
}

func (con *CContext) GetLastType() string {
	outBuf := *con.OutBuf
	if len(outBuf) == 0 {
		return ""
	}
	return outBuf[len(outBuf)-1].Type
}

func (con *CContext) GetLastBody() string {
	outBuf := *con.OutBuf
	if len(outBuf) == 0 {
		return ""
	}
	return outBuf[len(outBuf)-1].Body
}

func (con *CContext) SetLastBody(newBody string) error {
	outBuf := *con.OutBuf
	if len(outBuf) == 0 {
		return errors.New("outbuf is empty")
	}
	outBuf[len(outBuf)-1].Body = newBody
	return nil
}

func (con *CContext) GetLastTemplate() string {
	outBuf := *con.OutBuf
	if len(outBuf) == 0 {
		return ""
	}
	return outBuf[len(outBuf)-1].TemplateName
}

type Fragment struct {
	Body         string
	TemplateName string
	Index        int
}

// nolint
type CTemplateSet struct {
	templateList          map[string]*parse.Tree
	fileDir               string
	funcMap               map[string]interface{}
	importMap             map[string]string
	TemplateFragmentCount map[string]int
	Fragments             map[string]int
	fragmentCursor        map[string]int
	FragOut               string
	fragBuf               []Fragment
	varList               map[string]VarItem
	localVars             map[string]map[string]VarItemReflect
	hasDispInt            bool
	localDispStructIndex  int
	langIndexToName       []string
	stats                 map[string]int
	previousNode          parse.NodeType
	currentNode           parse.NodeType
	nextNode              parse.NodeType
	//tempVars map[string]string
	config        CTemplateConfig
	baseImportMap map[string]string
	buildTags     string
	expectsInt    interface{}
}

func NewCTemplateSet() *CTemplateSet {
	return &CTemplateSet{
		config: CTemplateConfig{
			PackageName: "main",
		},
		baseImportMap: map[string]string{},
		funcMap: map[string]interface{}{
			"and":      "&&",
			"not":      "!",
			"or":       "||",
			"eq":       "==",
			"ge":       ">=",
			"gt":       ">",
			"le":       "<=",
			"lt":       "<",
			"ne":       "!=",
			"add":      "+",
			"subtract": "-",
			"multiply": "*",
			"divide":   "/",
			"dock":     true,
			"elapsed":  true,
			"lang":     true,
			"level":    true,
			"scope":    true,
			"dyntmpl":  true,
		},
	}
}

func (c *CTemplateSet) SetConfig(config CTemplateConfig) {
	if config.PackageName == "" {
		config.PackageName = "main"
	}
	c.config = config
}

func (c *CTemplateSet) GetConfig() CTemplateConfig {
	return c.config
}

func (c *CTemplateSet) SetBaseImportMap(importMap map[string]string) {
	c.baseImportMap = importMap
}

func (c *CTemplateSet) SetBuildTags(tags string) {
	c.buildTags = tags
}

func (c *CTemplateSet) Compile(name string, fileDir string, expects string, expectsInt interface{}, varList map[string]VarItem, imports ...string) (out string, err error) {
	if c.config.Debug {
		fmt.Println("Compiling template '" + name + "'")
	}
	c.importMap = map[string]string{}
	for index, item := range c.baseImportMap {
		c.importMap[index] = item
	}
	if len(imports) > 0 {
		for _, importItem := range imports {
			c.importMap[importItem] = importItem
		}
	}

	c.fileDir = fileDir
	c.varList = varList
	c.hasDispInt = false
	c.localDispStructIndex = 0
	c.stats = make(map[string]int)
	c.expectsInt = expectsInt

	res, err := ioutil.ReadFile(fileDir + "overrides/" + name)
	if err != nil {
		c.detail("override path: ", fileDir+"overrides/"+name)
		c.detail("override err: ", err)
		res, err = ioutil.ReadFile(fileDir + name)
		if err != nil {
			return "", err
		}
	}
	content := string(res)
	if c.config.Minify {
		content = minify(content)
	}

	tree := parse.New(name, c.funcMap)
	var treeSet = make(map[string]*parse.Tree)
	tree, err = tree.Parse(content, "{{", "}}", treeSet, c.funcMap)
	if err != nil {
		return "", err
	}
	c.detail(name)

	fname := strings.TrimSuffix(name, filepath.Ext(name))
	var outBuf []OutBufferFrame
	con := CContext{VarHolder: "tmpl_" + fname + "_vars", HoldReflect: reflect.ValueOf(expectsInt), TemplateName: fname, OutBuf: &outBuf}
	c.templateList = map[string]*parse.Tree{fname: tree}
	c.detail(c.templateList)
	c.localVars = make(map[string]map[string]VarItemReflect)
	c.localVars[fname] = make(map[string]VarItemReflect)
	c.localVars[fname]["."] = VarItemReflect{".", con.VarHolder, con.HoldReflect}
	if c.Fragments == nil {
		c.Fragments = make(map[string]int)
	}
	c.fragmentCursor = map[string]int{fname: 0}
	c.langIndexToName = nil

	// TODO: Is this the first template loaded in? We really should have some sort of constructor for CTemplateSet
	if c.TemplateFragmentCount == nil {
		c.TemplateFragmentCount = make(map[string]int)
	}

	c.rootIterate(c.templateList[fname], con)
	c.TemplateFragmentCount[fname] = c.fragmentCursor[fname] + 1

	if len(c.langIndexToName) > 0 {
		c.importMap[langPkg] = langPkg
	}

	var importList string
	for _, item := range c.importMap {
		importList += "import \"" + item + "\"\n"
	}

	var varString string
	for _, varItem := range c.varList {
		varString += "var " + varItem.Name + " " + varItem.Type + " = " + varItem.Destination + "\n"
	}

	var fout string
	if c.buildTags != "" {
		fout += "// +build " + c.buildTags + "\n\n"
	}

	fout += "// Code generated by Gosora. More below:\n/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */\n"
	fout += "package " + c.config.PackageName + "\n" + importList + "\n"

	if !c.config.SkipInitBlock {
		if len(c.langIndexToName) > 0 {
			fout += "var " + fname + "_tmpl_phrase_id int\n\n"
		}
		fout += "// nolint\nfunc init() {\n"

		if !c.config.SkipHandles {
			fout += "\tcommon.Template_" + fname + "_handle = Template_" + fname + "\n"
			fout += "\tcommon.Ctemplates = append(common.Ctemplates,\"" + fname + "\")\n\tcommon.TmplPtrMap[\"" + fname + "\"] = &common.Template_" + fname + "_handle\n"
		}

		if !c.config.SkipTmplPtrMap {
			fout += "\tcommon.TmplPtrMap[\"o_" + fname + "\"] = Template_" + fname + "\n"
		}
		if len(c.langIndexToName) > 0 {
			fout += "\t" + fname + "_tmpl_phrase_id = phrases.RegisterTmplPhraseNames([]string{\n"
			for _, name := range c.langIndexToName {
				fout += "\t\t" + `"` + name + `"` + ",\n"
			}
			fout += "\t})\n"
		}
		fout += "}\n\n"
	}

	fout += "// nolint\nfunc Template_" + fname + "(tmpl_" + fname + "_vars " + expects + ", w io.Writer) error {\n"
	if len(c.langIndexToName) > 0 {
		fout += "var plist = phrases.GetTmplPhrasesBytes(" + fname + "_tmpl_phrase_id)\n"
	}
	fout += varString
	for _, frame := range outBuf {
		fout += frame.Body
	}
	fout += "return nil\n}\n"

	fout = strings.Replace(fout, `))
w.Write([]byte(`, " + ", -1)
	fout = strings.Replace(fout, "` + `", "", -1)

	for _, frag := range c.fragBuf {
		fragmentPrefix := frag.TemplateName + "_frags[" + strconv.Itoa(frag.Index) + "]"
		c.FragOut += fragmentPrefix + " = []byte(`" + frag.Body + "`)\n"
	}

	if c.config.Debug {
		for index, count := range c.stats {
			fmt.Println(index+": ", strconv.Itoa(count))
		}
		fmt.Println(" ")
	}
	c.detail("Output!")
	c.detail(fout)
	return fout, nil
}

func (c *CTemplateSet) rootIterate(tree *parse.Tree, con CContext) {
	c.detail(tree.Root)
	treeLength := len(tree.Root.Nodes)
	for index, node := range tree.Root.Nodes {
		c.detail("Node:", node.String())
		c.previousNode = c.currentNode
		c.currentNode = node.Type()
		if treeLength != (index + 1) {
			c.nextNode = tree.Root.Nodes[index+1].Type()
		}
		c.compileSwitch(con, node)
	}
}

func (c *CTemplateSet) compileSwitch(con CContext, node parse.Node) {
	c.dumpCall("compileSwitch", con, node)
	switch node := node.(type) {
	case *parse.ActionNode:
		c.detail("Action Node")
		if node.Pipe == nil {
			break
		}
		for _, cmd := range node.Pipe.Cmds {
			c.compileSubSwitch(con, cmd)
		}
	case *parse.IfNode:
		c.detail("If Node:")
		c.detail("node.Pipe", node.Pipe)
		var expr string
		for _, cmd := range node.Pipe.Cmds {
			c.detail("If Node Bit:", cmd)
			c.detail("Bit Type:", reflect.ValueOf(cmd).Type().Name())
			expr += c.compileExprSwitch(con, cmd)
			c.detail("Expression Step:", c.compileExprSwitch(con, cmd))
		}

		c.detail("Expression:", expr)
		c.previousNode = c.currentNode
		c.currentNode = parse.NodeList
		c.nextNode = -1
		con.Push("startif", "if "+expr+" {\n")
		c.compileSwitch(con, node.List)
		if node.ElseList == nil {
			c.detail("Selected Branch 1")
			con.Push("endif", "}\n")
		} else {
			c.detail("Selected Branch 2")
			con.Push("endif", "}")
			con.Push("startelse", " else {\n")
			c.compileSwitch(con, node.ElseList)
			con.Push("endelse", "}\n")
		}
	case *parse.ListNode:
		c.detail("List Node")
		for _, subnode := range node.Nodes {
			c.compileSwitch(con, subnode)
		}
	case *parse.RangeNode:
		c.compileRangeNode(con, node)
	case *parse.TemplateNode:
		c.compileSubTemplate(con, node)
	case *parse.TextNode:
		c.previousNode = c.currentNode
		c.currentNode = node.Type()
		c.nextNode = 0
		tmpText := bytes.TrimSpace(node.Text)
		if len(tmpText) == 0 {
			return
		}

		fragmentName := con.TemplateName + "_" + strconv.Itoa(c.fragmentCursor[con.TemplateName])
		fragmentPrefix := con.TemplateName + "_frags[" + strconv.Itoa(c.fragmentCursor[con.TemplateName]) + "]"
		_, ok := c.Fragments[fragmentName]
		if !ok {
			c.Fragments[fragmentName] = len(node.Text)
			c.fragBuf = append(c.fragBuf, Fragment{string(node.Text), con.TemplateName, c.fragmentCursor[con.TemplateName]})
		}
		c.fragmentCursor[con.TemplateName] = c.fragmentCursor[con.TemplateName] + 1
		con.Push("text", "w.Write("+fragmentPrefix+")\n")
	default:
		c.unknownNode(node)
	}
}

func (c *CTemplateSet) compileRangeNode(con CContext, node *parse.RangeNode) {
	c.dumpCall("compileRangeNode", con, node)
	c.detail("node.Pipe: ", node.Pipe)
	var expr string
	var outVal reflect.Value
	for _, cmd := range node.Pipe.Cmds {
		c.detail("Range Bit:", cmd)
		// ! This bit is slightly suspect, hm.
		expr, outVal = c.compileReflectSwitch(con, cmd)
	}
	c.detail("Expr:", expr)
	c.detail("Range Kind Switch!")

	var startIf = func(item reflect.Value, useCopy bool) {
		con.Push("startif", "if len("+expr+") != 0 {\n")
		con.Push("startloop", "for _, item := range "+expr+" {\n")
		ccon := con
		ccon.VarHolder = "item"
		ccon.HoldReflect = item
		c.compileSwitch(ccon, node.List)
		con.Push("endloop", "}\n")
		if node.ElseList != nil {
			con.Push("endif", "}")
			con.Push("startelse", " else {\n")
			if !useCopy {
				ccon = con
			}
			c.compileSwitch(ccon, node.ElseList)
			con.Push("endelse", "}\n")
		} else {
			con.Push("endloop", "}\n")
		}
	}

	switch outVal.Kind() {
	case reflect.Map:
		var item reflect.Value
		for _, key := range outVal.MapKeys() {
			item = outVal.MapIndex(key)
		}
		c.detail("Range item:", item)
		if !item.IsValid() {
			panic("item" + "^\n" + "Invalid map. Maybe, it doesn't have any entries for the template engine to analyse?")
		}
		startIf(item, true)
	case reflect.Slice:
		if outVal.Len() == 0 {
			panic("The sample data needs at-least one or more elements for the slices. We're looking into removing this requirement at some point!")
		}
		startIf(outVal.Index(0), false)
	case reflect.Invalid:
		return
	}
}

func (c *CTemplateSet) compileSubSwitch(con CContext, node *parse.CommandNode) {
	c.dumpCall("compileSubSwitch", con, node)
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
	case *parse.FieldNode:
		c.detail("Field Node:", n.Ident)
		/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Variable declarations are coming soon! */
		cur := con.HoldReflect

		var varBit string
		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
			varBit += ".(" + cur.Type().Name() + ")"
		}

		// ! Might not work so well for non-struct pointers
		skipPointers := func(cur reflect.Value, id string) reflect.Value {
			if cur.Kind() == reflect.Ptr {
				c.detail("Looping over pointer")
				for cur.Kind() == reflect.Ptr {
					cur = cur.Elem()
				}
				c.detail("Data Kind:", cur.Kind().String())
				c.detail("Field Bit:", id)
			}
			return cur
		}

		var assLines string
		var multiline = false
		for _, id := range n.Ident {
			c.detail("Data Kind:", cur.Kind().String())
			c.detail("Field Bit:", id)
			cur = skipPointers(cur, id)

			if !cur.IsValid() {
				c.error("Debug Data:")
				c.error("Holdreflect:", con.HoldReflect)
				c.error("Holdreflect.Kind():", con.HoldReflect.Kind())
				if !c.config.SuperDebug {
					c.error("cur.Kind():", cur.Kind().String())
				}
				c.error("")
				if !multiline {
					panic(con.VarHolder + varBit + "^\n" + "Invalid value. Maybe, it doesn't exist?")
				}
				panic(varBit + "^\n" + "Invalid value. Maybe, it doesn't exist?")
			}

			c.detail("in-loop varBit: " + varBit)
			if cur.Kind() == reflect.Map {
				cur = cur.MapIndex(reflect.ValueOf(id))
				varBit += "[\"" + id + "\"]"
				cur = skipPointers(cur, id)

				if cur.Kind() == reflect.Struct || cur.Kind() == reflect.Interface {
					// TODO: Move the newVarByte declaration to the top level or to the if level, if a dispInt is only used in a particular if statement
					var dispStr, newVarByte string
					if cur.Kind() == reflect.Interface {
						dispStr = "Int"
						if !c.hasDispInt {
							newVarByte = ":"
							c.hasDispInt = true
						}
					}
					// TODO: De-dupe identical struct types rather than allocating a variable for each one
					if cur.Kind() == reflect.Struct {
						dispStr = "Struct" + strconv.Itoa(c.localDispStructIndex)
						newVarByte = ":"
						c.localDispStructIndex++
					}
					con.VarHolder = "disp" + dispStr
					varBit = con.VarHolder + " " + newVarByte + "= " + con.VarHolder + varBit + "\n"
					multiline = true
				} else {
					continue
				}
			}
			if cur.Kind() != reflect.Interface {
				cur = cur.FieldByName(id)
				varBit += "." + id
			}

			// TODO: Handle deeply nested pointers mixed with interfaces mixed with pointers better
			if cur.Kind() == reflect.Interface {
				cur = cur.Elem()
				varBit += ".("
				// TODO: Surely, there's a better way of doing this?
				if cur.Type().PkgPath() != "main" && cur.Type().PkgPath() != "" {
					c.importMap["html/template"] = "html/template"
					varBit += strings.TrimPrefix(cur.Type().PkgPath(), "html/") + "."
				}
				varBit += cur.Type().Name() + ")"
			}
			c.detail("End Cycle: ", varBit)
		}

		if multiline {
			assSplit := strings.Split(varBit, "\n")
			varBit = assSplit[len(assSplit)-1]
			assSplit = assSplit[:len(assSplit)-1]
			assLines = strings.Join(assSplit, "\n") + "\n"
		}
		c.compileVarSub(con, con.VarHolder+varBit, cur, assLines, func(in string) string {
			for _, varItem := range c.varList {
				if strings.HasPrefix(in, varItem.Destination) {
					in = strings.Replace(in, varItem.Destination, varItem.Name, 1)
				}
			}
			return in
		})
	case *parse.DotNode:
		c.detail("Dot Node:", node.String())
		c.compileVarSub(con, con.VarHolder, con.HoldReflect, "", nil)
	case *parse.NilNode:
		panic("Nil is not a command x.x")
	case *parse.VariableNode:
		c.detail("Variable Node:", n.String())
		c.detail(n.Ident)
		varname, reflectVal := c.compileIfVarSub(con, n.String())
		c.compileVarSub(con, varname, reflectVal, "", nil)
	case *parse.StringNode:
		con.Push("stringnode", n.Quoted)
	case *parse.IdentifierNode:
		c.detail("Identifier Node:", node)
		c.detail("Identifier Node Args:", node.Args)
		out, outval, lit := c.compileIdentSwitch(con, node)
		if lit {
			con.Push("identifier", out)
			return
		}
		c.compileVarSub(con, out, outval, "", nil)
	default:
		c.unknownNode(node)
	}
}

func (c *CTemplateSet) compileExprSwitch(con CContext, node *parse.CommandNode) (out string) {
	c.dumpCall("compileExprSwitch", con, node)
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
	case *parse.FieldNode:
		if c.config.SuperDebug {
			fmt.Println("Field Node:", n.Ident)
			for _, id := range n.Ident {
				fmt.Println("Field Bit:", id)
			}
		}
		/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Coming Soon. */
		out = c.compileBoolSub(con, n.String())
	case *parse.ChainNode:
		c.detail("Chain Node:", n.Node)
		c.detail("Node Args:", node.Args)
	case *parse.IdentifierNode:
		c.detail("Identifier Node:", node)
		c.detail("Node Args:", node.Args)
		out = c.compileIdentSwitchN(con, node)
	case *parse.DotNode:
		out = con.VarHolder
	case *parse.VariableNode:
		c.detail("Variable Node:", n.String())
		c.detail("Node Identifier:", n.Ident)
		out, _ = c.compileIfVarSub(con, n.String())
	case *parse.NilNode:
		panic("Nil is not a command x.x")
	case *parse.PipeNode:
		c.detail("Pipe Node!")
		c.detail(n)
		c.detail("Node Args:", node.Args)
		out += c.compileIdentSwitchN(con, node)
	default:
		c.unknownNode(firstWord)
	}
	c.retCall("compileExprSwitch", out)
	return out
}

func (c *CTemplateSet) unknownNode(node parse.Node) {
	fmt.Println("Unknown Kind:", reflect.ValueOf(node).Elem().Kind())
	fmt.Println("Unknown Type:", reflect.ValueOf(node).Elem().Type().Name())
	panic("I don't know what node this is! Grr...")
}

func (c *CTemplateSet) compileIdentSwitchN(con CContext, node *parse.CommandNode) (out string) {
	c.detail("in compileIdentSwitchN")
	out, _, _ = c.compileIdentSwitch(con, node)
	return out
}

func (c *CTemplateSet) dumpSymbol(pos int, node *parse.CommandNode, symbol string) {
	c.detail("symbol: ", symbol)
	c.detail("node.Args[pos + 1]", node.Args[pos+1])
	c.detail("node.Args[pos + 2]", node.Args[pos+2])
}

func (c *CTemplateSet) compareFunc(con CContext, pos int, node *parse.CommandNode, compare string) (out string) {
	c.dumpSymbol(pos, node, compare)
	return c.compileIfVarSubN(con, node.Args[pos+1].String()) + " " + compare + " " + c.compileIfVarSubN(con, node.Args[pos+2].String())
}

func (c *CTemplateSet) simpleMath(con CContext, pos int, node *parse.CommandNode, symbol string) (out string, val reflect.Value) {
	leftParam, val2 := c.compileIfVarSub(con, node.Args[pos+1].String())
	rightParam, val3 := c.compileIfVarSub(con, node.Args[pos+2].String())

	if val2.IsValid() {
		val = val2
	} else if val3.IsValid() {
		val = val3
	} else {
		// TODO: What does this do?
		numSample := 1
		val = reflect.ValueOf(numSample)
	}
	c.dumpSymbol(pos, node, symbol)
	return leftParam + " " + symbol + " " + rightParam, val
}

func (c *CTemplateSet) compareJoin(con CContext, pos int, node *parse.CommandNode, symbol string) (pos2 int, out string) {
	c.detailf("Building %s function", symbol)
	if pos == 0 {
		fmt.Println("pos:", pos)
		panic(symbol + " is missing a left operand")
	}
	if len(node.Args) <= pos {
		fmt.Println("post pos:", pos)
		fmt.Println("len(node.Args):", len(node.Args))
		panic(symbol + " is missing a right operand")
	}

	left := c.compileBoolSub(con, node.Args[pos-1].String())
	_, funcExists := c.funcMap[node.Args[pos+1].String()]

	var right string
	if !funcExists {
		right = c.compileBoolSub(con, node.Args[pos+1].String())
	}
	out = left + " " + symbol + " " + right

	c.detail("Left operand:", node.Args[pos-1])
	c.detail("Right operand:", node.Args[pos+1])
	if !funcExists {
		pos++
	}
	c.detail("pos:", pos)
	c.detail("len(node.Args):", len(node.Args))

	return pos, out
}

func (c *CTemplateSet) compileIdentSwitch(con CContext, node *parse.CommandNode) (out string, val reflect.Value, literal bool) {
	c.dumpCall("compileIdentSwitch", con, node)
	var litString = func(inner string) {
		out = "w.Write([]byte(" + inner + "))\n"
		literal = true
	}
ArgLoop:
	for pos := 0; pos < len(node.Args); pos++ {
		id := node.Args[pos]
		c.detail("pos:", pos)
		c.detail("ID:", id)
		switch id.String() {
		case "not":
			out += "!"
		case "or", "and":
			var rout string
			pos, rout = c.compareJoin(con, pos, node, c.funcMap[id.String()].(string)) // TODO: Test this
			out += rout
		case "le", "lt", "gt", "ge", "eq", "ne":
			out += c.compareFunc(con, pos, node, c.funcMap[id.String()].(string))
			break ArgLoop
		case "add", "subtract", "divide", "multiply":
			rout, rval := c.simpleMath(con, pos, node, c.funcMap[id.String()].(string))
			out += rout
			val = rval
			break ArgLoop
		case "elapsed":
			leftOperand := node.Args[pos+1].String()
			leftParam, _ := c.compileIfVarSub(con, leftOperand)
			// TODO: Refactor this
			// TODO: Validate that this is actually a time.Time
			litString("time.Since(" + leftParam + ").String()")
			c.importMap["time"] = "time"
			break ArgLoop
		case "dock":
			// TODO: Implement string literals properly
			leftOperand := node.Args[pos+1].String()
			rightOperand := node.Args[pos+2].String()
			if len(leftOperand) == 0 || len(rightOperand) == 0 {
				panic("The left or right operand for function dock cannot be left blank")
			}

			leftParam := leftOperand
			if leftOperand[0] != '"' {
				leftParam, _ = c.compileIfVarSub(con, leftParam)
			}
			if rightOperand[0] == '"' {
				panic("The right operand for function dock cannot be a string")
			}
			rightParam, val3 := c.compileIfVarSub(con, rightOperand)
			if !val3.IsValid() {
				panic("val3 is invalid")
			}
			val = val3

			// TODO: Refactor this
			litString("common.BuildWidget(" + leftParam + "," + rightParam + ")")
			break ArgLoop
		case "lang":
			// TODO: Implement string literals properly
			leftOperand := node.Args[pos+1].String()
			if len(leftOperand) == 0 {
				panic("The left operand for the language string cannot be left blank")
			}
			if leftOperand[0] != '"' {
				panic("Phrase names cannot be dynamic")
			}

			// ! Slightly crude but it does the job
			leftParam := strings.Replace(leftOperand, "\"", "", -1)
			c.langIndexToName = append(c.langIndexToName, leftParam)
			litString("plist[" + strconv.Itoa(len(c.langIndexToName)-1) + "]")
			break ArgLoop
		case "level":
			// TODO: Implement level literals
			leftOperand := node.Args[pos+1].String()
			if len(leftOperand) == 0 {
				panic("The leftoperand for function level cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOperand)
			// TODO: Refactor this
			litString("phrases.GetLevelPhrase(" + leftParam + ")")
			c.importMap[langPkg] = langPkg
			break ArgLoop
		case "scope":
			literal = true
			break ArgLoop
		case "dyntmpl":
			var pageParam, headParam string
			// TODO: Implement string literals properly
			// TODO: Should we check to see if pos+3 is within the bounds of the slice?
			nameOperand := node.Args[pos+1].String()
			pageOperand := node.Args[pos+2].String()
			headOperand := node.Args[pos+3].String()
			if len(nameOperand) == 0 || len(pageOperand) == 0 || len(headOperand) == 0 {
				panic("None of the three operands for function dyntmpl can be left blank")
			}
			nameParam := nameOperand
			if nameOperand[0] != '"' {
				nameParam, _ = c.compileIfVarSub(con, nameParam)
			}
			if pageOperand[0] == '"' {
				panic("The page operand for function dyntmpl cannot be a string")
			}
			if headOperand[0] == '"' {
				panic("The head operand for function dyntmpl cannot be a string")
			}

			pageParam, val3 := c.compileIfVarSub(con, pageOperand)
			if !val3.IsValid() {
				panic("val3 is invalid")
			}
			headParam, val4 := c.compileIfVarSub(con, headOperand)
			if !val4.IsValid() {
				panic("val4 is invalid")
			}
			val = val4

			// TODO: Refactor this
			// TODO: Call the template function directly rather than going through RunThemeTemplate to eliminate a round of indirection?
			out = "{\nerr := common.RunThemeTemplate(" + headParam + ".Theme.Name," + nameParam + "," + pageParam + ",w)\n"
			out += "if err != nil {\nreturn err\n}\n}\n"
			literal = true
			break ArgLoop
		default:
			c.detail("Variable!")
			if len(node.Args) > (pos + 1) {
				nextNode := node.Args[pos+1].String()
				if nextNode == "or" || nextNode == "and" {
					continue
				}
			}
			out += c.compileIfVarSubN(con, id.String())
		}
	}
	c.retCall("compileIdentSwitch", out, val, literal)
	return out, val, literal
}

func (c *CTemplateSet) compileReflectSwitch(con CContext, node *parse.CommandNode) (out string, outVal reflect.Value) {
	c.dumpCall("compileReflectSwitch", con, node)
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
	case *parse.FieldNode:
		if c.config.SuperDebug {
			fmt.Println("Field Node:", n.Ident)
			for _, id := range n.Ident {
				fmt.Println("Field Bit:", id)
			}
		}
		/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Coming Soon. */
		return c.compileIfVarSub(con, n.String())
	case *parse.ChainNode:
		c.detail("Chain Node:", n.Node)
		c.detail("node.Args:", node.Args)
	case *parse.DotNode:
		return con.VarHolder, con.HoldReflect
	case *parse.NilNode:
		panic("Nil is not a command x.x")
	default:
		//panic("I don't know what node this is")
	}
	return out, outVal
}

func (c *CTemplateSet) compileIfVarSubN(con CContext, varname string) (out string) {
	c.dumpCall("compileIfVarSubN", con, varname)
	out, _ = c.compileIfVarSub(con, varname)
	return out
}

func (c *CTemplateSet) compileIfVarSub(con CContext, varname string) (out string, val reflect.Value) {
	c.dumpCall("compileIfVarSub", con, varname)
	cur := con.HoldReflect
	if varname[0] != '.' && varname[0] != '$' {
		return varname, cur
	}

	var stepInterface = func() {
		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
			out += ".(" + cur.Type().Name() + ")"
		}
	}
	bits := strings.Split(varname, ".")
	if varname[0] == '$' {
		var res VarItemReflect
		if varname[1] == '.' {
			res = c.localVars[con.TemplateName]["."]
		} else {
			res = c.localVars[con.TemplateName][strings.TrimPrefix(bits[0], "$")]
		}
		out += res.Destination
		cur = res.Value

		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
		}
	} else {
		out += con.VarHolder
		stepInterface()
	}
	bits[0] = strings.TrimPrefix(bits[0], "$")

	var dumpKind = func(pre string) {
		c.detail(pre+" Kind:", cur.Kind())
		c.detail(pre+" Type:", cur.Type().Name())
	}
	dumpKind("Cur")
	for _, bit := range bits {
		c.detail("Variable Field:", bit)
		if bit == "" {
			continue
		}

		// TODO: Fix this up so that it works for regular pointers and not just struct pointers. Ditto for the other cur.Kind() == reflect.Ptr we have in this file
		if cur.Kind() == reflect.Ptr {
			c.detail("Looping over pointer")
			for cur.Kind() == reflect.Ptr {
				cur = cur.Elem()
			}
			c.detail("Data Kind:", cur.Kind().String())
			c.detail("Field Bit:", bit)
		}

		cur = cur.FieldByName(bit)
		out += "." + bit
		stepInterface()
		if !cur.IsValid() {
			fmt.Println("cur: ", cur)
			panic(out + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		dumpKind("Data")
	}

	c.detail("Out Value:", out)
	dumpKind("Out")
	for _, varItem := range c.varList {
		if strings.HasPrefix(out, varItem.Destination) {
			out = strings.Replace(out, varItem.Destination, varItem.Name, 1)
		}
	}

	_, ok := c.stats[out]
	if ok {
		c.stats[out]++
	} else {
		c.stats[out] = 1
	}

	c.retCall("compileIfVarSub", out, cur)
	return out, cur
}

func (c *CTemplateSet) compileBoolSub(con CContext, varname string) string {
	c.dumpCall("compileBoolSub", con, varname)
	out, val := c.compileIfVarSub(con, varname)
	// TODO: What if it's a pointer or an interface? I *think* we've got pointers handled somewhere, but not interfaces which we don't know the types of at compile time
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		out += " > 0"
	case reflect.Bool: // Do nothing
	case reflect.String:
		out += " != \"\""
	case reflect.Slice, reflect.Map:
		out = "len(" + out + ") != 0"
	default:
		fmt.Println("Variable Name:", varname)
		fmt.Println("Variable Holder:", con.VarHolder)
		fmt.Println("Variable Kind:", con.HoldReflect.Kind())
		panic("I don't know what this variable's type is o.o\n")
	}
	c.retCall("compileBoolSub", out)
	return out
}

// For debugging the template generator
func (c *CTemplateSet) debugParam(param interface{}, depth int) (pstr string) {
	switch p := param.(type) {
	case CContext:
		return "con,"
	case reflect.Value:
		if p.Kind() == reflect.Ptr || p.Kind() == reflect.Interface {
			for p.Kind() == reflect.Ptr || p.Kind() == reflect.Interface {
				if p.Kind() == reflect.Ptr {
					pstr += "*"
				} else {
					pstr += "£"
				}
				p = p.Elem()
			}
		}
		kind := p.Kind().String()
		if kind != "struct" {
			pstr += kind
		} else {
			pstr += p.Type().Name()
		}
		return pstr + ","
	case string:
		return "\"" + p + "\","
	case int:
		return strconv.Itoa(p) + ","
	case bool:
		if p {
			return "true,"
		}
		return "false,"
	case func(string) string:
		if p == nil {
			return "nil,"
		}
		return "func(string) string),"
	default:
		return "?,"
	}
}
func (c *CTemplateSet) dumpCall(name string, params ...interface{}) {
	var pstr string
	for _, param := range params {
		pstr += c.debugParam(param, 0)
	}
	if len(pstr) > 0 {
		pstr = pstr[:len(pstr)-1]
	}
	c.detail("called " + name + "(" + pstr + ")")
}
func (c *CTemplateSet) retCall(name string, params ...interface{}) {
	var pstr string
	for _, param := range params {
		pstr += c.debugParam(param, 0)
	}
	if len(pstr) > 0 {
		pstr = pstr[:len(pstr)-1]
	}
	c.detail("returned from " + name + " => (" + pstr + ")")
}

func (c *CTemplateSet) compileVarSub(con CContext, varname string, val reflect.Value, assLines string, onEnd func(string) string) {
	c.dumpCall("compileVarSub", con, varname, val, assLines, onEnd)
	if onEnd == nil {
		onEnd = func(in string) string {
			return in
		}
	}

	// Is this a literal string?
	if len(varname) != 0 && varname[0] == '"' {
		con.Push("varsub", onEnd(assLines+"w.Write([]byte("+varname+"))\n"))
		return
	}
	for _, varItem := range c.varList {
		if strings.HasPrefix(varname, varItem.Destination) {
			varname = strings.Replace(varname, varItem.Destination, varItem.Name, 1)
		}
	}

	_, ok := c.stats[varname]
	if ok {
		c.stats[varname]++
	} else {
		c.stats[varname] = 1
	}
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() == reflect.Ptr {
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
			varname = "*" + varname
		}
	}

	c.detail("varname: ", varname)
	c.detail("assLines: ", assLines)
	var base string
	switch val.Kind() {
	case reflect.Int:
		c.importMap["strconv"] = "strconv"
		base = "w.Write([]byte(strconv.Itoa(" + varname + ")))\n"
	case reflect.Bool:
		base = "if " + varname + " {\nw.Write([]byte(\"true\"))} else {\nw.Write([]byte(\"false\"))\n}\n"
	case reflect.String:
		if val.Type().Name() != "string" && !strings.HasPrefix(varname, "string(") {
			varname = "string(" + varname + ")"
		}
		base = "w.Write([]byte(" + varname + "))\n"
	case reflect.Int64:
		c.importMap["strconv"] = "strconv"
		base = "w.Write([]byte(strconv.FormatInt(" + varname + ", 10)))\n"
	default:
		if !val.IsValid() {
			panic(assLines + varname + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		fmt.Println("Unknown Variable Name:", varname)
		fmt.Println("Unknown Kind:", val.Kind())
		fmt.Println("Unknown Type:", val.Type().Name())
		panic("-- I don't know what this variable's type is o.o\n")
	}
	c.detail("base: ", base)
	con.Push("varsub", onEnd(assLines+base))
}

func (c *CTemplateSet) compileSubTemplate(pcon CContext, node *parse.TemplateNode) {
	c.detail("in compileSubTemplate")
	c.detail("Template Node: ", node.Name)

	fname := strings.TrimSuffix(node.Name, filepath.Ext(node.Name))
	con := pcon
	con.VarHolder = "tmpl_" + fname + "_vars"
	con.TemplateName = fname
	if node.Pipe != nil {
		for _, cmd := range node.Pipe.Cmds {
			firstWord := cmd.Args[0]
			switch firstWord.(type) {
			case *parse.DotNode:
				con.VarHolder = pcon.VarHolder
				con.HoldReflect = pcon.HoldReflect
			case *parse.NilNode:
				panic("Nil is not a command x.x")
			default:
				c.detail("Unknown Node: ", firstWord)
				panic("")
			}
		}
	}

	// TODO: Cascade errors back up the tree to the caller?
	res, err := ioutil.ReadFile(c.fileDir + "overrides/" + node.Name)
	if err != nil {
		c.detail("override path: ", c.fileDir+"overrides/"+node.Name)
		c.detail("override err: ", err)
		res, err = ioutil.ReadFile(c.fileDir + node.Name)
		if err != nil {
			log.Fatal(err)
		}
	}
	content := string(res)
	if c.config.Minify {
		content = minify(content)
	}

	tree := parse.New(node.Name, c.funcMap)
	var treeSet = make(map[string]*parse.Tree)
	tree, err = tree.Parse(content, "{{", "}}", treeSet, c.funcMap)
	if err != nil {
		log.Fatal(err)
	}

	c.templateList[fname] = tree
	subtree := c.templateList[fname]
	c.detail("subtree.Root", subtree.Root)

	c.localVars[fname] = make(map[string]VarItemReflect)
	c.localVars[fname]["."] = VarItemReflect{".", con.VarHolder, con.HoldReflect}
	c.fragmentCursor[fname] = 0
	c.rootIterate(subtree, con)
	c.TemplateFragmentCount[fname] = c.fragmentCursor[fname] + 1
}

// TODO: Should we rethink the way the log methods work or their names?

func (c *CTemplateSet) detail(args ...interface{}) {
	if c.config.SuperDebug {
		fmt.Println(args...)
	}
}

func (c *CTemplateSet) detailf(left string, args ...interface{}) {
	if c.config.SuperDebug {
		fmt.Printf(left, args...)
	}
}

func (c *CTemplateSet) error(args ...interface{}) {
	if c.config.Debug {
		fmt.Println(args...)
	}
}
