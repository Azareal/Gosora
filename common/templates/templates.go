package tmpl

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template/parse"
	"time"
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

// nolint
type CTemplateSet struct {
	templateList map[string]*parse.Tree
	fileDir      string
	funcMap      map[string]interface{}
	importMap    map[string]string
	//templateFragmentCount map[string]int
	fragOnce             map[string]bool
	fragmentCursor       map[string]int
	FragOut              []OutFrag
	fragBuf              []Fragment
	varList              map[string]VarItem
	localVars            map[string]map[string]VarItemReflect
	hasDispInt           bool
	localDispStructIndex int
	langIndexToName      []string
	guestOnly            bool
	memberOnly           bool
	stats                map[string]int
	//tempVars map[string]string
	config        CTemplateConfig
	baseImportMap map[string]string
	buildTags     string

	overridenTrack map[string]map[string]bool
	overridenRoots map[string]map[string]bool
	themeName      string
	perThemeTmpls  map[string]bool

	logger *log.Logger
}

func NewCTemplateSet(in string) *CTemplateSet {
	f, err := os.OpenFile("./logs/tmpls-"+in+"-"+strconv.FormatInt(time.Now().Unix(), 10)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	return &CTemplateSet{
		config: CTemplateConfig{
			PackageName: "main",
		},
		baseImportMap:  map[string]string{},
		overridenRoots: map[string]map[string]bool{},
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
			//"langf":true,
			"level":   true,
			"abstime": true,
			"reltime": true,
			"scope":   true,
			"dyntmpl": true,
			"index":   true,
		},
		logger: log.New(f, "", log.LstdFlags),
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

func (c *CTemplateSet) SetOverrideTrack(overriden map[string]map[string]bool) {
	c.overridenTrack = overriden
}

func (c *CTemplateSet) GetOverridenRoots() map[string]map[string]bool {
	return c.overridenRoots
}

func (c *CTemplateSet) SetThemeName(name string) {
	c.themeName = name
}

func (c *CTemplateSet) SetPerThemeTmpls(perThemeTmpls map[string]bool) {
	c.perThemeTmpls = perThemeTmpls
}

func (c *CTemplateSet) ResetLogs(in string) {
	f, err := os.OpenFile("./logs/tmpls-"+in+"-"+strconv.FormatInt(time.Now().Unix(), 10)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	c.logger = log.New(f, "", log.LstdFlags)
}

type SkipBlock struct {
	Frags           map[int]int
	LastCount       int
	ClosestFragSkip int
}
type Skipper struct {
	Count int
	Index int
}

type OutFrag struct {
	TmplName string
	Index    int
	Body     string
}

func (c *CTemplateSet) CompileByLoggedin(name string, fileDir string, expects string, expectsInt interface{}, varList map[string]VarItem, imports ...string) (stub string, gout string, mout string, err error) {
	c.importMap = map[string]string{}
	for index, item := range c.baseImportMap {
		c.importMap[index] = item
	}
	for _, importItem := range imports {
		c.importMap[importItem] = importItem
	}
	var importList string
	for _, item := range c.importMap {
		importList += "import \"" + item + "\"\n"
	}

	fname := strings.TrimSuffix(name, filepath.Ext(name))
	if c.themeName != "" {
		_, ok := c.perThemeTmpls[fname]
		if !ok {
			return "", "", "", nil
		}
		fname += "_" + c.themeName
	}
	c.importMap["github.com/Azareal/Gosora/common"] = "github.com/Azareal/Gosora/common"

	stub = `package ` + c.config.PackageName + `
` + importList + `
import "errors"
`

	if !c.config.SkipInitBlock {
		stub += "// nolint\nfunc init() {\n"

		if !c.config.SkipHandles && c.themeName == "" {
			stub += "\tcommon.Template_" + fname + "_handle = Template_" + fname + "\n"
			stub += "\tcommon.Ctemplates = append(common.Ctemplates,\"" + fname + "\")\n"
		}

		if !c.config.SkipTmplPtrMap {
			stub += "tmpl := Template_" + fname + "\n"
			stub += "\tcommon.TmplPtrMap[\"" + fname + "\"] = &tmpl\n"
			stub += "\tcommon.TmplPtrMap[\"o_" + fname + "\"] = tmpl\n"
		}

		stub += "}\n\n"
	}

	// TODO: Try to remove this redundant interface cast
	stub += `
// nolint
func Template_` + fname + `(tmpl_` + fname + `_i interface{}, w io.Writer) error {
	tmpl_` + fname + `_vars, ok := tmpl_` + fname + `_i.(` + expects + `)
	if !ok {
		return errors.New("invalid page struct value")
	}
	if tmpl_` + fname + `_vars.CurrentUser.Loggedin {
		return Template_` + fname + `_member(tmpl_` + fname + `_i, w)
	}
	return Template_` + fname + `_guest(tmpl_` + fname + `_i, w)
}`

	c.fileDir = fileDir
	content, err := c.loadTemplate(c.fileDir, name)
	if err != nil {
		c.detail("bailing out: ", err)
		return "", "", "", err
	}

	c.guestOnly = true
	gout, err = c.compile(name, content, expects, expectsInt, varList, imports...)
	if err != nil {
		return "", "", "", err
	}
	c.guestOnly = false

	c.memberOnly = true
	mout, err = c.compile(name, content, expects, expectsInt, varList, imports...)
	c.memberOnly = false

	return stub, gout, mout, err
}

func (c *CTemplateSet) Compile(name string, fileDir string, expects string, expectsInt interface{}, varList map[string]VarItem, imports ...string) (out string, err error) {
	if c.config.Debug {
		c.logger.Println("Compiling template '" + name + "'")
	}
	c.fileDir = fileDir
	content, err := c.loadTemplate(c.fileDir, name)
	if err != nil {
		c.detail("bailing out: ", err)
		return "", err
	}

	return c.compile(name, content, expects, expectsInt, varList, imports...)
}

func (c *CTemplateSet) compile(name string, content string, expects string, expectsInt interface{}, varList map[string]VarItem, imports ...string) (out string, err error) {
	//c.dumpCall("compile", name, content, expects, expectsInt, varList, imports)
	//c.detailf("c: %+v\n", c)
	c.importMap = map[string]string{}
	for index, item := range c.baseImportMap {
		c.importMap[index] = item
	}
	c.importMap["errors"] = "errors"
	for _, importItem := range imports {
		c.importMap[importItem] = importItem
	}

	c.varList = varList
	c.hasDispInt = false
	c.localDispStructIndex = 0
	c.stats = make(map[string]int)

	tree := parse.New(name, c.funcMap)
	var treeSet = make(map[string]*parse.Tree)
	tree, err = tree.Parse(content, "{{", "}}", treeSet, c.funcMap)
	if err != nil {
		return "", err
	}
	c.detail(name)

	fname := strings.TrimSuffix(name, filepath.Ext(name))
	if c.themeName != "" {
		_, ok := c.perThemeTmpls[fname]
		if !ok {
			c.detail("fname not in c.perThemeTmpls")
			c.detail("c.perThemeTmpls", c.perThemeTmpls)
			return "", nil
		}
		fname += "_" + c.themeName
	}
	if c.guestOnly {
		fname += "_guest"
	} else if c.memberOnly {
		fname += "_member"
	}

	c.detail("root overridenTrack loop")
	c.detail("fname:", fname)
	for themeName, track := range c.overridenTrack {
		c.detail("themeName:", themeName)
		c.detailf("track: %+v\n", track)
		croot, ok := c.overridenRoots[themeName]
		if !ok {
			croot = make(map[string]bool)
			c.overridenRoots[themeName] = croot
		}
		c.detailf("croot: %+v\n", croot)
		for tmplName, _ := range track {
			cname := tmplName
			if c.guestOnly {
				cname += "_guest"
			} else if c.memberOnly {
				cname += "_member"
			}
			c.detail("cname:", cname)
			if fname == cname {
				c.detail("match")
				croot[strings.TrimSuffix(strings.TrimSuffix(fname, "_guest"), "_member")] = true
			} else {
				c.detail("no match")
			}
		}
	}
	c.detailf("c.overridenRoots: %+v\n", c.overridenRoots)

	var outBuf []OutBufferFrame
	var rootHold = "tmpl_" + fname + "_vars"
	con := CContext{
		RootHolder:       rootHold,
		VarHolder:        rootHold,
		HoldReflect:      reflect.ValueOf(expectsInt),
		RootTemplateName: fname,
		TemplateName:     fname,
		OutBuf:           &outBuf,
	}
	c.templateList = map[string]*parse.Tree{fname: tree}
	c.detail(c.templateList)
	c.localVars = make(map[string]map[string]VarItemReflect)
	c.localVars[fname] = make(map[string]VarItemReflect)
	c.localVars[fname]["."] = VarItemReflect{".", con.VarHolder, con.HoldReflect}
	if c.fragOnce == nil {
		c.fragOnce = make(map[string]bool)
	}
	c.fragmentCursor = map[string]int{fname: 0}
	c.fragBuf = nil
	c.langIndexToName = nil

	// TODO: Is this the first template loaded in? We really should have some sort of constructor for CTemplateSet
	//if c.templateFragmentCount == nil {
	//	c.templateFragmentCount = make(map[string]int)
	//}
	//c.detailf("c: %+v\n", c)

	startIndex := con.StartTemplate("")
	c.rootIterate(c.templateList[fname], con)
	con.EndTemplate("")
	c.afterTemplate(con, startIndex)
	//c.templateFragmentCount[fname] = c.fragmentCursor[fname] + 1

	_, ok := c.fragOnce[fname]
	if !ok {
		c.fragOnce[fname] = true
	}
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

		if !c.config.SkipHandles && c.themeName == "" {
			fout += "\tcommon.Template_" + fname + "_handle = Template_" + fname + "\n"
			fout += "\tcommon.Ctemplates = append(common.Ctemplates,\"" + fname + "\")\n"
		}

		if !c.config.SkipTmplPtrMap {
			fout += "tmpl := Template_" + fname + "\n"
			fout += "\tcommon.TmplPtrMap[\"" + fname + "\"] = &tmpl\n"
			fout += "\tcommon.TmplPtrMap[\"o_" + fname + "\"] = tmpl\n"
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

	fout += "// nolint\nfunc Template_" + fname + "(tmpl_" + fname + "_i interface{}, w io.Writer) error {\n"
	fout += `tmpl_` + fname + `_vars, ok := tmpl_` + fname + `_i.(` + expects + `)
	if !ok {
		return errors.New("invalid page struct value")
	}
`
	if len(c.langIndexToName) > 0 {
		fout += "var plist = phrases.GetTmplPhrasesBytes(" + fname + "_tmpl_phrase_id)\n"
	}
	fout += varString

	var skipped = make(map[string]*SkipBlock) // map[templateName]*SkipBlock{map[atIndexAndAfter]skipThisMuch,lastCount}

	var writeTextFrame = func(tmplName string, index int) {
		out := "w.Write(" + tmplName + "_frags[" + strconv.Itoa(index) + "]" + ")\n"
		c.detail("writing ", out)
		fout += out
	}

	for fid := 0; len(outBuf) > fid; fid++ {
		frame := outBuf[fid]
		c.detail(frame.Type + " frame")
		if frame.Type == "text" {
			c.detail(frame)
			oid := fid
			c.detail("oid:", oid)
			skipBlock, ok := skipped[frame.TemplateName]
			if !ok {
				skipBlock = &SkipBlock{make(map[int]int), 0, 0}
				skipped[frame.TemplateName] = skipBlock
			}
			skip := skipBlock.LastCount
			c.detailf("skipblock %+v\n", skipBlock)
			//var count int
			for len(outBuf) > fid+1 && outBuf[fid+1].Type == "text" && outBuf[fid+1].TemplateName == frame.TemplateName {
				c.detail("pre fid:", fid)
				//count++
				next := outBuf[fid+1]
				c.detail("next frame:", next)
				c.detail("frame frag:", c.fragBuf[frame.Extra2.(int)])
				c.detail("next frag:", c.fragBuf[next.Extra2.(int)])
				c.fragBuf[frame.Extra2.(int)].Body += c.fragBuf[next.Extra2.(int)].Body
				c.fragBuf[next.Extra2.(int)].Seen = true
				fid++
				skipBlock.LastCount++
				skipBlock.Frags[frame.Extra.(int)] = skipBlock.LastCount
				c.detail("post fid:", fid)
			}
			writeTextFrame(frame.TemplateName, frame.Extra.(int)-skip)
		} else if frame.Type == "varsub" || frame.Type == "cvarsub" {
			fout += "w.Write(" + frame.Body + ")\n"
		} else if frame.Type == "identifier" {
			fout += frame.Body
		} else if frame.Type == "lang" {
			fout += "w.Write(plist[" + strconv.Itoa(frame.Extra.(int)) + "])\n"
		} else {
			fout += frame.Body
		}
	}
	fout += "return nil\n}\n"

	var writeFrag = func(tmplName string, index int, body string) {
		//c.detail("writing ", fragmentPrefix)
		c.FragOut = append(c.FragOut, OutFrag{tmplName, index, body})
	}

	for _, frag := range c.fragBuf {
		c.detail("frag: ", frag)
		if frag.Seen {
			c.detail("invisible")
			continue
		}
		// TODO: What if the same template is invoked in multiple spots in a template?
		skipBlock := skipped[frag.TemplateName]
		skip := skipBlock.Frags[skipBlock.ClosestFragSkip]
		_, ok := skipBlock.Frags[frag.Index]
		if ok {
			skipBlock.ClosestFragSkip = frag.Index
		}
		c.detailf("skipblock %+v\n", skipBlock)
		c.detail("skipping ", skip)
		index := frag.Index - skip
		if index < 0 {
			index = 0
		}
		writeFrag(frag.TemplateName, index, frag.Body)
	}

	fout = strings.Replace(fout, `))
w.Write([]byte(`, " + ", -1)
	fout = strings.Replace(fout, "` + `", "", -1)

	if c.config.Debug {
		for index, count := range c.stats {
			c.logger.Println(index+": ", strconv.Itoa(count))
		}
		c.logger.Println(" ")
	}
	c.detail("Output!")
	c.detail(fout)
	return fout, nil
}

func (c *CTemplateSet) rootIterate(tree *parse.Tree, con CContext) {
	c.dumpCall("rootIterate", tree, con)
	c.detail(tree.Root)
	for _, node := range tree.Root.Nodes {
		c.detail("Node:", node.String())
		c.compileSwitch(con, node)
	}
	c.retCall("rootIterate")
}

func inSlice(haystack []string, expr string) bool {
	for _, needle := range haystack {
		if needle == expr {
			return true
		}
	}
	return false
}

func (c *CTemplateSet) compileSwitch(con CContext, node parse.Node) {
	c.dumpCall("compileSwitch", con, node)
	defer c.retCall("compileSwitch")
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
		// Simple member / guest optimisation for now
		// TODO: Expand upon this
		var userExprs = []string{
			con.RootHolder + ".CurrentUser.Loggedin",
			con.RootHolder + ".CurrentUser.IsSuperMod",
			con.RootHolder + ".CurrentUser.IsAdmin",
		}
		var negUserExprs = []string{
			"!" + con.RootHolder + ".CurrentUser.Loggedin",
			"!" + con.RootHolder + ".CurrentUser.IsSuperMod",
			"!" + con.RootHolder + ".CurrentUser.IsAdmin",
		}
		if c.guestOnly {
			c.detail("optimising away member branch")
			if inSlice(userExprs, expr) {
				c.detail("positive conditional:", expr)
				if node.ElseList != nil {
					c.compileSwitch(con, node.ElseList)
				}
				return
			} else if inSlice(negUserExprs, expr) {
				c.detail("negative conditional:", expr)
				c.compileSwitch(con, node.List)
				return
			}
		} else if c.memberOnly {
			c.detail("optimising away guest branch")
			if (con.RootHolder + ".CurrentUser.Loggedin") == expr {
				c.detail("positive conditional:", expr)
				c.compileSwitch(con, node.List)
				return
			} else if ("!" + con.RootHolder + ".CurrentUser.Loggedin") == expr {
				c.detail("negative conditional:", expr)
				if node.ElseList != nil {
					c.compileSwitch(con, node.ElseList)
				}
				return
			}
		}

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
		c.detailf("List Node: %+v\n", node)
		for _, subnode := range node.Nodes {
			c.compileSwitch(con, subnode)
		}
	case *parse.RangeNode:
		c.compileRangeNode(con, node)
	case *parse.TemplateNode:
		c.compileSubTemplate(con, node)
	case *parse.TextNode:
		c.addText(con, node.Text)
	default:
		c.unknownNode(node)
	}
}

func (c *CTemplateSet) addText(con CContext, text []byte) {
	c.dumpCall("addText", con, text)
	tmpText := bytes.TrimSpace(text)
	if len(tmpText) == 0 {
		return
	}
	nodeText := string(text)
	c.detail("con.TemplateName: ", con.TemplateName)
	fragIndex := c.fragmentCursor[con.TemplateName]
	_, ok := c.fragOnce[con.TemplateName]
	c.fragBuf = append(c.fragBuf, Fragment{nodeText, con.TemplateName, fragIndex, ok})
	con.PushText(strconv.Itoa(fragIndex), fragIndex, len(c.fragBuf)-1)
	c.fragmentCursor[con.TemplateName] = fragIndex + 1
}

func (c *CTemplateSet) compileRangeNode(con CContext, node *parse.RangeNode) {
	c.dumpCall("compileRangeNode", con, node)
	defer c.retCall("compileRangeNode")
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
		startIndex := con.StartLoop("for _, item := range " + expr + " {\n")
		ccon := con
		var depth string
		if ccon.VarHolder == "item" {
			depth = strings.TrimPrefix(ccon.VarHolder, "item")
			if depth != "" {
				idepth, err := strconv.Atoi(depth)
				if err != nil {
					panic(err)
				}
				depth = strconv.Itoa(idepth + 1)
			}
		}
		ccon.VarHolder = "item" + depth
		ccon.HoldReflect = item
		c.compileSwitch(ccon, node.List)
		if con.LastBufIndex() == startIndex {
			con.DiscardAndAfter(startIndex - 1)
			return
		}
		con.EndLoop("}\n")
		c.afterTemplate(con, startIndex)
		if node.ElseList != nil {
			con.Push("endif", "}")
			con.Push("startelse", " else {\n")
			if !useCopy {
				ccon = con
			}
			c.compileSwitch(ccon, node.ElseList)
			con.Push("endelse", "}\n")
		} else {
			con.Push("endif", "}\n")
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
			c.critical("expr:", expr)
			c.critical("con.VarHolder", con.VarHolder)
			panic("item" + "^\n" + "Invalid map. Maybe, it doesn't have any entries for the template engine to analyse?")
		}
		startIf(item, true)
	case reflect.Slice:
		if outVal.Len() == 0 {
			c.critical("expr:", expr)
			c.critical("con.VarHolder", con.VarHolder)
			panic("The sample data needs at-least one or more elements for the slices. We're looking into removing this requirement at some point!")
		}
		startIf(outVal.Index(0), false)
	case reflect.Invalid:
		return
	}
}

// ! Temporary, we probably want something that is good with non-struct pointers too
// For compileSubSwitch and compileSubTemplate
func (c *CTemplateSet) skipStructPointers(cur reflect.Value, id string) reflect.Value {
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

// For compileSubSwitch and compileSubTemplate
func (c *CTemplateSet) checkIfValid(cur reflect.Value, varHolder string, holdreflect reflect.Value, varBit string, multiline bool) {
	if !cur.IsValid() {
		c.critical("Debug Data:")
		c.critical("Holdreflect:", holdreflect)
		c.critical("Holdreflect.Kind():", holdreflect.Kind())
		if !c.config.SuperDebug {
			c.critical("cur.Kind():", cur.Kind().String())
		}
		c.critical("")
		if !multiline {
			panic(varHolder + varBit + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		panic(varBit + "^\n" + "Invalid value. Maybe, it doesn't exist?")
	}
}

func (c *CTemplateSet) compileSubSwitch(con CContext, node *parse.CommandNode) {
	c.dumpCall("compileSubSwitch", con, node)
	switch n := node.Args[0].(type) {
	case *parse.FieldNode:
		c.detail("Field Node:", n.Ident)
		/* Use reflect to determine if the field is for a method, otherwise assume it's a variable. Variable declarations are coming soon! */
		cur := con.HoldReflect

		var varBit string
		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
			varBit += ".(" + cur.Type().Name() + ")"
		}

		var assLines string
		var multiline = false
		for _, id := range n.Ident {
			c.detail("Data Kind:", cur.Kind().String())
			c.detail("Field Bit:", id)
			cur = c.skipStructPointers(cur, id)
			c.checkIfValid(cur, con.VarHolder, con.HoldReflect, varBit, multiline)

			c.detail("in-loop varBit: " + varBit)
			if cur.Kind() == reflect.Map {
				cur = cur.MapIndex(reflect.ValueOf(id))
				varBit += "[\"" + id + "\"]"
				cur = c.skipStructPointers(cur, id)

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
		out, outval, lit, noident := c.compileIdentSwitch(con, node)
		if noident {
			return
		} else if lit {
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
			c.logger.Println("Field Node:", n.Ident)
			for _, id := range n.Ident {
				c.logger.Println("Field Bit:", id)
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
	c.logger.Println("Unknown Kind:", reflect.ValueOf(node).Elem().Kind())
	c.logger.Println("Unknown Type:", reflect.ValueOf(node).Elem().Type().Name())
	panic("I don't know what node this is! Grr...")
}

func (c *CTemplateSet) compileIdentSwitchN(con CContext, node *parse.CommandNode) (out string) {
	c.detail("in compileIdentSwitchN")
	out, _, _, _ = c.compileIdentSwitch(con, node)
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
		c.logger.Println("pos:", pos)
		panic(symbol + " is missing a left operand")
	}
	if len(node.Args) <= pos {
		c.logger.Println("post pos:", pos)
		c.logger.Println("len(node.Args):", len(node.Args))
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

func (c *CTemplateSet) compileIdentSwitch(con CContext, node *parse.CommandNode) (out string, val reflect.Value, literal bool, notident bool) {
	c.dumpCall("compileIdentSwitch", con, node)
	var litString = func(inner string, bytes bool) {
		if !bytes {
			inner = "StringToBytes(" + inner + ")"
		}
		out = "w.Write(" + inner + ")\n"
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
			litString("time.Since("+leftParam+").String()", false)
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
			litString("common.BuildWidget("+leftParam+","+rightParam+")", false)
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
			notident = true
			con.PushPhrase(len(c.langIndexToName) - 1)
			break ArgLoop
		// TODO: Implement langf
		case "level":
			// TODO: Implement level literals
			leftOperand := node.Args[pos+1].String()
			if len(leftOperand) == 0 {
				panic("The leftoperand for function level cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOperand)
			// TODO: Refactor this
			litString("phrases.GetLevelPhrase("+leftParam+")", false)
			c.importMap[langPkg] = langPkg
			break ArgLoop
		case "abstime":
			// TODO: Implement level literals
			leftOperand := node.Args[pos+1].String()
			if len(leftOperand) == 0 {
				panic("The leftoperand for function abstime cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOperand)
			// TODO: Refactor this
			litString(leftParam+".Format(\"2006-01-02 15:04:05\")", false)
			break ArgLoop
		case "reltime":
			// TODO: Implement level literals
			leftOperand := node.Args[pos+1].String()
			if len(leftOperand) == 0 {
				panic("The leftoperand for function reltime cannot be left blank")
			}
			leftParam, _ := c.compileIfVarSub(con, leftOperand)
			// TODO: Refactor this
			litString("common.RelativeTime("+leftParam+")", false)
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
			out = "{\nerr := " + headParam + ".Theme.RunTmpl(" + nameParam + "," + pageParam + ",w)\n"
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
	return out, val, literal, notident
}

func (c *CTemplateSet) compileReflectSwitch(con CContext, node *parse.CommandNode) (out string, outVal reflect.Value) {
	c.dumpCall("compileReflectSwitch", con, node)
	firstWord := node.Args[0]
	switch n := firstWord.(type) {
	case *parse.FieldNode:
		if c.config.SuperDebug {
			c.logger.Println("Field Node:", n.Ident)
			for _, id := range n.Ident {
				c.logger.Println("Field Bit:", id)
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
		var nobreak = (cur.Type().Name() == "nobreak")
		c.detailf("cur.Type().Name(): %+v\n", cur.Type().Name())
		if cur.Kind() == reflect.Interface && !nobreak {
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
		if !cur.IsValid() {
			c.logger.Println("cur: ", cur)
			panic(out + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		stepInterface()
		if !cur.IsValid() {
			c.logger.Println("cur: ", cur)
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
		c.logger.Println("Variable Name:", varname)
		c.logger.Println("Variable Holder:", con.VarHolder)
		c.logger.Println("Variable Kind:", con.HoldReflect.Kind())
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
	defer c.retCall("compileVarSub")
	if onEnd == nil {
		onEnd = func(in string) string {
			return in
		}
	}

	// Is this a literal string?
	if len(varname) != 0 && varname[0] == '"' {
		con.Push("lvarsub", onEnd(assLines+"w.Write(StringToBytes("+varname+"))\n"))
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
		base = "StringToBytes(strconv.Itoa(" + varname + "))"
	case reflect.Bool:
		// TODO: Take c.memberOnly into account
		// TODO: Make this a template fragment so more optimisations can be applied to this
		// TODO: De-duplicate this logic
		var userExprs = []string{
			con.RootHolder + ".CurrentUser.Loggedin",
			con.RootHolder + ".CurrentUser.IsSuperMod",
			con.RootHolder + ".CurrentUser.IsAdmin",
		}
		var negUserExprs = []string{
			"!" + con.RootHolder + ".CurrentUser.Loggedin",
			"!" + con.RootHolder + ".CurrentUser.IsSuperMod",
			"!" + con.RootHolder + ".CurrentUser.IsAdmin",
		}
		if c.guestOnly {
			c.detail("optimising away member branch")
			if inSlice(userExprs, varname) {
				c.detail("positive conditional:", varname)
				c.addText(con, []byte("false"))
				return
			} else if inSlice(negUserExprs, varname) {
				c.detail("negative conditional:", varname)
				c.addText(con, []byte("true"))
				return
			}
		} else if c.memberOnly {
			c.detail("optimising away guest branch")
			if (con.RootHolder + ".CurrentUser.Loggedin") == varname {
				c.detail("positive conditional:", varname)
				c.addText(con, []byte("true"))
				return
			} else if ("!" + con.RootHolder + ".CurrentUser.Loggedin") == varname {
				c.detail("negative conditional:", varname)
				c.addText(con, []byte("false"))
				return
			}
		}
		con.Push("startif", "if "+varname+" {\n")
		c.addText(con, []byte("true"))
		con.Push("endif", "} ")
		con.Push("startelse", "else {\n")
		c.addText(con, []byte("false"))
		con.Push("endelse", "}\n")
		return
	case reflect.String:
		if val.Type().Name() != "string" && !strings.HasPrefix(varname, "string(") {
			varname = "string(" + varname + ")"
		}
		base = "StringToBytes(" + varname + ")"
		// We don't to waste time on this conversion / w.Write call when guests don't have sessions
		// TODO: Implement this properly
		if c.guestOnly && base == "StringToBytes("+con.RootHolder+".CurrentUser.Session))" {
			return
		}
	case reflect.Int64:
		c.importMap["strconv"] = "strconv"
		base = "StringToBytes(strconv.FormatInt(" + varname + ", 10))"
	case reflect.Struct:
		// TODO: Avoid clashing with other packages which have structs named Time
		if val.Type().Name() == "Time" {
			base = "StringToBytes(" + varname + ".String())"
		} else {
			if !val.IsValid() {
				panic(assLines + varname + "^\n" + "Invalid value. Maybe, it doesn't exist?")
			}
			c.logger.Println("Unknown Struct Name:", varname)
			c.logger.Println("Unknown Struct:", val.Type().Name())
			panic("-- I don't know what this variable's type is o.o\n")
		}
	default:
		if !val.IsValid() {
			panic(assLines + varname + "^\n" + "Invalid value. Maybe, it doesn't exist?")
		}
		c.logger.Println("Unknown Variable Name:", varname)
		c.logger.Println("Unknown Kind:", val.Kind())
		c.logger.Println("Unknown Type:", val.Type().Name())
		panic("-- I don't know what this variable's type is o.o\n")
	}
	c.detail("base: ", base)
	if assLines == "" {
		con.Push("varsub", base)
	} else {
		con.Push("lvarsub", onEnd(assLines+base))
	}
}

func (c *CTemplateSet) compileSubTemplate(pcon CContext, node *parse.TemplateNode) {
	c.dumpCall("compileSubTemplate", pcon, node)
	defer c.retCall("compileSubTemplate")
	c.detail("Template Node: ", node.Name)

	// TODO: Cascade errors back up the tree to the caller?
	content, err := c.loadTemplate(c.fileDir, node.Name)
	if err != nil {
		c.logger.Fatal(err)
	}

	tree := parse.New(node.Name, c.funcMap)
	var treeSet = make(map[string]*parse.Tree)
	tree, err = tree.Parse(content, "{{", "}}", treeSet, c.funcMap)
	if err != nil {
		c.logger.Fatal(err)
	}

	fname := strings.TrimSuffix(node.Name, filepath.Ext(node.Name))
	if c.themeName != "" {
		_, ok := c.perThemeTmpls[fname]
		if !ok {
			c.detail("fname not in c.perThemeTmpls")
			c.detail("c.perThemeTmpls", c.perThemeTmpls)
		}
		fname += "_" + c.themeName
	}
	if c.guestOnly {
		fname += "_guest"
	} else if c.memberOnly {
		fname += "_member"
	}

	con := pcon
	con.VarHolder = "tmpl_" + fname + "_vars"
	con.TemplateName = fname
	if node.Pipe != nil {
		for _, cmd := range node.Pipe.Cmds {
			switch p := cmd.Args[0].(type) {
			case *parse.FieldNode:
				// TODO: Incomplete but it should cover the basics
				cur := pcon.HoldReflect

				var varBit string
				if cur.Kind() == reflect.Interface {
					cur = cur.Elem()
					varBit += ".(" + cur.Type().Name() + ")"
				}

				for _, id := range p.Ident {
					c.detail("Data Kind:", cur.Kind().String())
					c.detail("Field Bit:", id)
					cur = c.skipStructPointers(cur, id)
					c.checkIfValid(cur, pcon.VarHolder, pcon.HoldReflect, varBit, false)

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
				}
				con.VarHolder = pcon.VarHolder + varBit
				con.HoldReflect = cur
			case *parse.DotNode:
				con.VarHolder = pcon.VarHolder
				con.HoldReflect = pcon.HoldReflect
			case *parse.NilNode:
				panic("Nil is not a command x.x")
			default:
				c.critical("Unknown Param Type:", p)
				pvar := reflect.ValueOf(p)
				c.critical("param kind:", pvar.Kind().String())
				c.critical("param type:", pvar.Type().Name())
				if pvar.Kind() == reflect.Ptr {
					c.critical("Looping over pointer")
					for pvar.Kind() == reflect.Ptr {
						pvar = pvar.Elem()
					}
					c.critical("concrete kind:", pvar.Kind().String())
					c.critical("concrete type:", pvar.Type().Name())
				}
				panic("")
			}
		}
	}

	c.templateList[fname] = tree
	subtree := c.templateList[fname]
	c.detail("subtree.Root", subtree.Root)
	c.localVars[fname] = make(map[string]VarItemReflect)
	c.localVars[fname]["."] = VarItemReflect{".", con.VarHolder, con.HoldReflect}
	c.fragmentCursor[fname] = 0

	var startBit, endBit string
	if con.LoopDepth != 0 {
		startBit = "{\n"
		endBit = "}\n"
	}
	con.StartTemplate(startBit)
	c.rootIterate(subtree, con)
	con.EndTemplate(endBit)
	//c.templateFragmentCount[fname] = c.fragmentCursor[fname] + 1

	_, ok := c.fragOnce[fname]
	if !ok {
		c.fragOnce[fname] = true
	}

	// map[string]map[string]bool
	c.detail("overridenTrack loop")
	c.detail("fname:", fname)
	for themeName, track := range c.overridenTrack {
		c.detail("themeName:", themeName)
		c.detailf("track: %+v\n", track)
		croot, ok := c.overridenRoots[themeName]
		if !ok {
			croot = make(map[string]bool)
			c.overridenRoots[themeName] = croot
		}
		c.detailf("croot: %+v\n", croot)
		for tmplName, _ := range track {
			cname := tmplName
			if c.guestOnly {
				cname += "_guest"
			} else if c.memberOnly {
				cname += "_member"
			}
			c.detail("cname:", cname)
			if fname == cname {
				c.detail("match")
				croot[strings.TrimSuffix(strings.TrimSuffix(con.RootTemplateName, "_guest"), "_member")] = true
			} else {
				c.detail("no match")
			}
		}
	}
	c.detailf("c.overridenRoots: %+v\n", c.overridenRoots)
}

func (c *CTemplateSet) loadTemplate(fileDir string, name string) (content string, err error) {
	c.dumpCall("loadTemplate", fileDir, name)

	c.detail("c.themeName: ", c.themeName)
	if c.themeName != "" {
		c.detail("per-theme override: ", "./themes/"+c.themeName+"/overrides/"+name)
		res, err := ioutil.ReadFile("./themes/" + c.themeName + "/overrides/" + name)
		if err == nil {
			content = string(res)
			if c.config.Minify {
				content = Minify(content)
			}
			return content, nil
		}
		c.detail("override err: ", err)
	}

	res, err := ioutil.ReadFile(c.fileDir + "overrides/" + name)
	if err != nil {
		c.detail("override path: ", c.fileDir+"overrides/"+name)
		c.detail("override err: ", err)
		res, err = ioutil.ReadFile(c.fileDir + name)
		if err != nil {
			return "", err
		}
	}
	content = string(res)
	if c.config.Minify {
		content = Minify(content)
	}
	return content, nil
}

func (c *CTemplateSet) afterTemplate(con CContext, startIndex int) {
	c.dumpCall("afterTemplate", con, startIndex)
	defer c.retCall("afterTemplate")

	var loopDepth = 0
	var outBuf = *con.OutBuf
	var varcounts = make(map[string]int)
	var loopStart = startIndex

	if outBuf[startIndex].Type == "startloop" && (len(outBuf) > startIndex+1) {
		loopStart++
	}

	// Exclude varsubs within loops for now
	for i := loopStart; i < len(outBuf); i++ {
		item := outBuf[i]
		c.detail("item:", item)
		if item.Type == "startloop" {
			loopDepth++
			c.detail("loopDepth:", loopDepth)
		} else if item.Type == "endloop" {
			loopDepth--
			c.detail("loopDepth:", loopDepth)
			if loopDepth == -1 {
				break
			}
		} else if item.Type == "varsub" && loopDepth == 0 {
			count := varcounts[item.Body]
			varcounts[item.Body] = count + 1
			c.detail("count " + strconv.Itoa(count) + " for " + item.Body)
			c.detail("loopDepth:", loopDepth)
		}
	}

	var varstr string
	var i int
	var varmap = make(map[string]int)
	for name, count := range varcounts {
		if count > 1 {
			varstr += "var cached_var_" + strconv.Itoa(i) + " = " + name + "\n"
			varmap[name] = i
			i++
		}
	}

	// Exclude varsubs within loops for now
	loopDepth = 0
	for i := loopStart; i < len(outBuf); i++ {
		item := outBuf[i]
		if item.Type == "startloop" {
			loopDepth++
		} else if item.Type == "endloop" {
			loopDepth--
			if loopDepth == -1 {
				break
			}
		} else if item.Type == "varsub" && loopDepth == 0 {
			index, ok := varmap[item.Body]
			if ok {
				item.Body = "cached_var_" + strconv.Itoa(index)
				item.Type = "cvarsub"
				outBuf[i] = item
			}
		}
	}

	con.AttachVars(varstr, startIndex)
}

// TODO: Should we rethink the way the log methods work or their names?

func (c *CTemplateSet) detail(args ...interface{}) {
	if c.config.SuperDebug {
		c.logger.Println(args...)
	}
}

func (c *CTemplateSet) detailf(left string, args ...interface{}) {
	if c.config.SuperDebug {
		c.logger.Printf(left, args...)
	}
}

func (c *CTemplateSet) error(args ...interface{}) {
	if c.config.Debug {
		c.logger.Println(args...)
	}
}

func (c *CTemplateSet) critical(args ...interface{}) {
	c.logger.Println(args...)
}
