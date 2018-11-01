package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/query_gen"
	"github.com/Azareal/Gosora/common/phrases"
)

type MenuItemList []MenuItem

type MenuListHolder struct {
	MenuID     int
	List       MenuItemList
	Variations map[int]menuTmpl // 0 = Guest Menu, 1 = Member Menu, 2 = Super Mod Menu, 3 = Admin Menu
}

type menuPath struct {
	Path  string
	Index int
}

type menuTmpl struct {
	RenderBuffer    [][]byte
	VariableIndices []int
	PathMappings    []menuPath
}

type MenuItem struct {
	ID     int
	MenuID int

	Name     string
	HTMLID   string
	CSSClass string
	Position string
	Path     string
	Aria     string
	Tooltip  string
	Order    int
	TmplName string

	GuestOnly    bool
	MemberOnly   bool
	SuperModOnly bool
	AdminOnly    bool
}

// TODO: Move the menu item stuff to it's own file
type MenuItemStmts struct {
	update      *sql.Stmt
	insert      *sql.Stmt
	delete      *sql.Stmt
	updateOrder *sql.Stmt
}

var menuItemStmts MenuItemStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		menuItemStmts = MenuItemStmts{
			update:      acc.Update("menu_items").Set("name = ?, htmlID = ?, cssClass = ?, position = ?, path = ?, aria = ?, tooltip = ?, tmplName = ?, guestOnly = ?, memberOnly = ?, staffOnly = ?, adminOnly = ?").Where("miid = ?").Prepare(),
			insert:      acc.Insert("menu_items").Columns("mid, name, htmlID, cssClass, position, path, aria, tooltip, tmplName, guestOnly, memberOnly, staffOnly, adminOnly").Fields("?,?,?,?,?,?,?,?,?,?,?,?,?").Prepare(),
			delete:      acc.Delete("menu_items").Where("miid = ?").Prepare(),
			updateOrder: acc.Update("menu_items").Set("order = ?").Where("miid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

func (item MenuItem) Commit() error {
	_, err := menuItemStmts.update.Exec(item.Name, item.HTMLID, item.CSSClass, item.Position, item.Path, item.Aria, item.Tooltip, item.TmplName, item.GuestOnly, item.MemberOnly, item.SuperModOnly, item.AdminOnly, item.ID)
	Menus.Load(item.MenuID)
	return err
}

func (item MenuItem) Create() (int, error) {
	res, err := menuItemStmts.insert.Exec(item.MenuID, item.Name, item.HTMLID, item.CSSClass, item.Position, item.Path, item.Aria, item.Tooltip, item.TmplName, item.GuestOnly, item.MemberOnly, item.SuperModOnly, item.AdminOnly)
	if err != nil {
		return 0, err
	}
	Menus.Load(item.MenuID)

	miid64, err := res.LastInsertId()
	return int(miid64), err
}

func (item MenuItem) Delete() error {
	_, err := menuItemStmts.delete.Exec(item.ID)
	Menus.Load(item.MenuID)
	return err
}

func (hold *MenuListHolder) LoadTmpl(name string) (menuTmpl MenuTmpl, err error) {
	data, err := ioutil.ReadFile("./templates/" + name + ".html")
	if err != nil {
		return menuTmpl, err
	}
	return hold.Parse(name, data), nil
}

// TODO: Make this atomic, maybe with a transaction or store the order on the menu itself?
func (hold *MenuListHolder) UpdateOrder(updateMap map[int]int) error {
	for miid, order := range updateMap {
		_, err := menuItemStmts.updateOrder.Exec(order, miid)
		if err != nil {
			return err
		}
	}
	Menus.Load(hold.MenuID)
	return nil
}

func (hold *MenuListHolder) LoadTmpls() (tmpls map[string]MenuTmpl, err error) {
	tmpls = make(map[string]MenuTmpl)
	var loadTmpl = func(name string) error {
		menuTmpl, err := hold.LoadTmpl(name)
		if err != nil {
			return err
		}
		tmpls[name] = menuTmpl
		return nil
	}
	err = loadTmpl("menu_item")
	if err != nil {
		return tmpls, err
	}
	err = loadTmpl("menu_alerts")
	return tmpls, err
}

// TODO: Run this in main, sync ticks, when the phrase file changes (need to implement the sync for that first), and when the settings are changed
func (hold *MenuListHolder) Preparse() error {
	tmpls, err := hold.LoadTmpls()
	if err != nil {
		return err
	}

	var addVariation = func(index int, callback func(mitem MenuItem) bool) {
		renderBuffer, variableIndices, pathList := hold.Scan(tmpls, callback)
		hold.Variations[index] = menuTmpl{renderBuffer, variableIndices, pathList}
	}

	// Guest Menu
	addVariation(0, func(mitem MenuItem) bool {
		return !mitem.MemberOnly
	})
	// Member Menu
	addVariation(1, func(mitem MenuItem) bool {
		return !mitem.SuperModOnly && !mitem.GuestOnly
	})
	// Super Mod Menu
	addVariation(2, func(mitem MenuItem) bool {
		return !mitem.AdminOnly && !mitem.GuestOnly
	})
	// Admin Menu
	addVariation(3, func(mitem MenuItem) bool {
		return !mitem.GuestOnly
	})
	return nil
}

func nextCharIs(tmplData []byte, i int, expects byte) bool {
	if len(tmplData) <= (i + 1) {
		return false
	}
	return tmplData[i+1] == expects
}

func peekNextChar(tmplData []byte, i int) byte {
	if len(tmplData) <= (i + 1) {
		return 0
	}
	return tmplData[i+1]
}

func skipUntilIfExists(tmplData []byte, i int, expects byte) (newI int, hasIt bool) {
	j := i
	for ; j < len(tmplData); j++ {
		if tmplData[j] == expects {
			return j, true
		}
	}
	return j, false
}

func skipUntilCharsExist(tmplData []byte, i int, expects []byte) (newI int, hasIt bool) {
	j := i
	expectIndex := 0
	for ; j < len(tmplData) && expectIndex < len(expects); j++ {
		//fmt.Println("tmplData[j]: ", string(tmplData[j]))
		if tmplData[j] != expects[expectIndex] {
			return j, false
		}
		//fmt.Printf("found %+v at %d\n", string(expects[expectIndex]), expectIndex)
		expectIndex++
	}
	return j, true
}

func skipAllUntilCharsExist(tmplData []byte, i int, expects []byte) (newI int, hasIt bool) {
	j := i
	expectIndex := 0
	for ; j < len(tmplData) && expectIndex < len(expects); j++ {
		if tmplData[j] == expects[expectIndex] {
			//fmt.Printf("expects[expectIndex]: %+v - %d\n", string(expects[expectIndex]), expectIndex)
			expectIndex++
			if len(expects) <= expectIndex {
				break
			}
		} else {
			/*if expectIndex != 0 {
				fmt.Println("broke expectations")
				fmt.Println("expected: ", string(expects[expectIndex]))
				fmt.Println("got: ", string(tmplData[j]))
				fmt.Println("next: ", string(peekNextChar(tmplData, j)))
				fmt.Println("next: ", string(peekNextChar(tmplData, j+1)))
				fmt.Println("next: ", string(peekNextChar(tmplData, j+2)))
				fmt.Println("next: ", string(peekNextChar(tmplData, j+3)))
			}*/
			expectIndex = 0
		}
	}
	return j, len(expects) == expectIndex
}

type menuRenderItem struct {
	Type  int // 0: text, 1: variable
	Index int
}

type MenuTmpl struct {
	Name           string
	TextBuffer     [][]byte
	VariableBuffer [][]byte
	RenderList     []menuRenderItem
}

func menuDumpSlice(outerSlice [][]byte) {
	for sliceID, slice := range outerSlice {
		fmt.Print(strconv.Itoa(sliceID) + ":[")
		for _, char := range slice {
			fmt.Print(string(char))
		}
		fmt.Print("] ")
	}
}

func (hold *MenuListHolder) Parse(name string, tmplData []byte) (menuTmpl MenuTmpl) {
	var textBuffer, variableBuffer [][]byte
	var renderList []menuRenderItem
	var subBuffer []byte

	// ? We only support simple properties on MenuItem right now
	var addVariable = func(name []byte) {
		// TODO: Check if the subBuffer has any items or is empty
		textBuffer = append(textBuffer, subBuffer)
		subBuffer = nil

		variableBuffer = append(variableBuffer, name)
		renderList = append(renderList, menuRenderItem{0, len(textBuffer) - 1})
		renderList = append(renderList, menuRenderItem{1, len(variableBuffer) - 1})
	}

	tmplData = bytes.Replace(tmplData, []byte("{{"), []byte("{"), -1)
	tmplData = bytes.Replace(tmplData, []byte("}}"), []byte("}}"), -1)
	for i := 0; i < len(tmplData); i++ {
		char := tmplData[i]
		if char == '{' {
			dotIndex, hasDot := skipUntilIfExists(tmplData, i, '.')
			if !hasDot {
				// Template function style
				langIndex, hasChars := skipUntilCharsExist(tmplData, i+1, []byte("lang"))
				if hasChars {
					startIndex, hasStart := skipUntilIfExists(tmplData, langIndex, '"')
					endIndex, hasEnd := skipUntilIfExists(tmplData, startIndex+1, '"')
					if hasStart && hasEnd {
						fenceIndex, hasFence := skipUntilIfExists(tmplData, endIndex, '}')
						if !hasFence || !nextCharIs(tmplData, fenceIndex, '}') {
							break
						}
						//fmt.Println("tmplData[startIndex:endIndex]: ", tmplData[startIndex+1:endIndex])
						prefix := []byte("lang.")
						addVariable(append(prefix, tmplData[startIndex+1:endIndex]...))
						i = fenceIndex + 1
						continue
					}
				}
				break
			}
			fenceIndex, hasFence := skipUntilIfExists(tmplData, dotIndex, '}')
			if !hasFence {
				break
			}
			addVariable(tmplData[dotIndex:fenceIndex])
			i = fenceIndex + 1
			continue
		}
		subBuffer = append(subBuffer, char)
	}
	if len(subBuffer) > 0 {
		// TODO: Have a property in renderList which holds the byte slice since variableBuffers and textBuffers have the same underlying implementation?
		textBuffer = append(textBuffer, subBuffer)
		renderList = append(renderList, menuRenderItem{0, len(textBuffer) - 1})
	}

	return MenuTmpl{name, textBuffer, variableBuffer, renderList}
}

func (hold *MenuListHolder) Scan(menuTmpls map[string]MenuTmpl, showItem func(mitem MenuItem) bool) (renderBuffer [][]byte, variableIndices []int, pathList []menuPath) {
	for _, mitem := range hold.List {
		// Do we want this item in this variation of the menu?
		if !showItem(mitem) {
			continue
		}
		renderBuffer, variableIndices = hold.ScanItem(menuTmpls, mitem, renderBuffer, variableIndices)
		pathList = append(pathList, menuPath{mitem.Path, len(renderBuffer) - 1})
	}

	// TODO: Need more coalescing in the renderBuffer
	return renderBuffer, variableIndices, pathList
}

// Note: This doesn't do a visibility check like hold.Scan() does
func (hold *MenuListHolder) ScanItem(menuTmpls map[string]MenuTmpl, mitem MenuItem, renderBuffer [][]byte, variableIndices []int) ([][]byte, []int) {
	menuTmpl, ok := menuTmpls[mitem.TmplName]
	if !ok {
		menuTmpl = menuTmpls["menu_item"]
	}

	for _, renderItem := range menuTmpl.RenderList {
		if renderItem.Type == 0 {
			renderBuffer = append(renderBuffer, menuTmpl.TextBuffer[renderItem.Index])
			continue
		}

		variable := menuTmpl.VariableBuffer[renderItem.Index]
		dotAt, hasDot := skipUntilIfExists(variable, 0, '.')
		if !hasDot {
			continue
		}

		if bytes.Equal(variable[:dotAt], []byte("lang")) {
			renderBuffer = append(renderBuffer, []byte(phrases.GetTmplPhrase(string(bytes.TrimPrefix(variable[dotAt:], []byte("."))))))
			continue
		}

		var renderItem []byte
		switch string(variable) {
		case ".ID":
			renderItem = []byte(strconv.Itoa(mitem.ID))
		case ".Name":
			renderItem = []byte(mitem.Name)
		case ".HTMLID":
			renderItem = []byte(mitem.HTMLID)
		case ".CSSClass":
			renderItem = []byte(mitem.CSSClass)
		case ".Position":
			renderItem = []byte(mitem.Position)
		case ".Path":
			renderItem = []byte(mitem.Path)
		case ".Aria":
			renderItem = []byte(mitem.Aria)
		case ".Tooltip":
			renderItem = []byte(mitem.Tooltip)
		case ".CSSActive":
			renderItem = []byte("{dyn.active}")
		}

		_, hasInnerVar := skipUntilIfExists(renderItem, 0, '{')
		if hasInnerVar {
			fmt.Println("inner var: ", string(renderItem))
			dotAt, hasDot := skipUntilIfExists(renderItem, 0, '.')
			endFence, hasEndFence := skipUntilIfExists(renderItem, dotAt, '}')
			if !hasDot || !hasEndFence || (endFence-dotAt) <= 1 {
				renderBuffer = append(renderBuffer, renderItem)
				variableIndices = append(variableIndices, len(renderBuffer)-1)
				continue
			}

			if bytes.Equal(renderItem[1:dotAt], []byte("lang")) {
				//fmt.Println("lang var: ", string(renderItem[dotAt+1:endFence]))
				renderBuffer = append(renderBuffer, []byte(phrases.GetTmplPhrase(string(renderItem[dotAt+1:endFence]))))
			} else {
				fmt.Println("other var: ", string(variable[:dotAt]))
				if len(renderItem) > 0 {
					renderBuffer = append(renderBuffer, renderItem)
					variableIndices = append(variableIndices, len(renderBuffer)-1)
				}
			}
			continue
		}
		if len(renderItem) > 0 {
			renderBuffer = append(renderBuffer, renderItem)
		}
	}
	return renderBuffer, variableIndices
}

// TODO: Pre-render the lang stuff
func (hold *MenuListHolder) Build(w io.Writer, user *User, pathPrefix string) error {
	var mTmpl menuTmpl
	if !user.Loggedin {
		mTmpl = hold.Variations[0]
	} else if user.IsAdmin {
		mTmpl = hold.Variations[3]
	} else if user.IsSuperMod {
		mTmpl = hold.Variations[2]
	} else {
		mTmpl = hold.Variations[1]
	}
	if pathPrefix == "" {
		pathPrefix = Config.DefaultPath
	}

	if len(mTmpl.VariableIndices) == 0 {
		for _, renderItem := range mTmpl.RenderBuffer {
			w.Write(renderItem)
		}
		return nil
	}

	var nearIndex = 0
	for index, renderItem := range mTmpl.RenderBuffer {
		if index != mTmpl.VariableIndices[nearIndex] {
			w.Write(renderItem)
			continue
		}
		variable := renderItem
		// ? - I can probably remove this check now that I've kicked it upstream, or we could keep it here for safety's sake?
		if len(variable) == 0 {
			continue
		}

		prevIndex := 0
		for i := 0; i < len(renderItem); i++ {
			fenceStart, hasFence := skipUntilIfExists(variable, i, '{')
			if !hasFence {
				continue
			}
			i = fenceStart
			fenceEnd, hasFence := skipUntilIfExists(variable, fenceStart, '}')
			if !hasFence {
				continue
			}
			i = fenceEnd
			dotAt, hasDot := skipUntilIfExists(variable, fenceStart, '.')
			if !hasDot {
				continue
			}

			switch string(variable[fenceStart+1 : dotAt]) {
			case "me":
				w.Write(variable[prevIndex:fenceStart])
				switch string(variable[dotAt+1 : fenceEnd]) {
				case "Link":
					w.Write([]byte(user.Link))
				case "Session":
					w.Write([]byte(user.Session))
				}
				prevIndex = fenceEnd
			// TODO: Optimise this
			case "dyn":
				w.Write(variable[prevIndex:fenceStart])
				var pmi int
				for ii, pathItem := range mTmpl.PathMappings {
					pmi = ii
					if pathItem.Index > index {
						break
					}
				}

				if len(mTmpl.PathMappings) != 0 {
					path := mTmpl.PathMappings[pmi].Path
					if path == "" || path == "/" {
						path = Config.DefaultPath
					}
					if strings.HasPrefix(path, pathPrefix) {
						w.Write([]byte(" menu_active"))
					}
				}

				prevIndex = fenceEnd
			}
		}

		w.Write(variable[prevIndex : len(variable)-1])
		if len(mTmpl.VariableIndices) > (nearIndex + 1) {
			nearIndex++
		}
	}
	return nil
}
