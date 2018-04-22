package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"sync/atomic"

	"../query_gen/lib"
)

var Menus *DefaultMenuStore

type MenuItemList []MenuItem

type DefaultMenuStore struct {
	menus map[int]*atomic.Value
}

func NewDefaultMenuStore() *DefaultMenuStore {
	return &DefaultMenuStore{make(map[int]*atomic.Value)}
}

func (store *DefaultMenuStore) Get(mid int) *MenuListHolder {
	aStore, ok := store.menus[mid]
	if ok {
		return aStore.Load().(*MenuListHolder)
	}
	return nil
}

type MenuListHolder struct {
	List       MenuItemList
	Variations map[int]menuTmpl // 0 = Guest Menu, 1 = Member Menu, 2 = Super Mod Menu, 3 = Admin Menu
}

type menuTmpl struct {
	RenderBuffer    [][]byte
	VariableIndices []int
}

type MenuItem struct {
	ID       int
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

func (store *DefaultMenuStore) Load(mid int) error {
	var mlist MenuItemList
	acc := qgen.Builder.Accumulator()
	err := acc.Select("menu_items").Columns("htmlID, cssClass, position, path, aria, tooltip, order, tmplName, guestOnly, memberOnly, staffOnly, adminOnly").Where("mid = " + strconv.Itoa(mid)).Orderby("order ASC").Each(func(rows *sql.Rows) error {
		var mitem = MenuItem{ID: 1}
		err := rows.Scan(&mitem.HTMLID, &mitem.CSSClass, &mitem.Position, &mitem.Path, &mitem.Aria, &mitem.Tooltip, &mitem.Order, &mitem.TmplName, &mitem.GuestOnly, &mitem.MemberOnly, &mitem.SuperModOnly, &mitem.AdminOnly)
		if err != nil {
			return err
		}
		mlist = append(mlist, mitem)
		return nil
	})
	if err != nil {
		return err
	}

	hold := &MenuListHolder{mlist, make(map[int]menuTmpl)}
	err = hold.Preparse()
	if err != nil {
		return err
	}

	var aStore = &atomic.Value{}
	aStore.Store(hold)
	store.menus[mid] = aStore
	return nil
}

// TODO: Run this in main, sync ticks, when the phrase file changes (need to implement the sync for that first), and when the settings are changed
func (hold *MenuListHolder) Preparse() error {
	var tmpls = make(map[string]MenuTmpl)
	var loadTmpl = func(name string) error {
		data, err := ioutil.ReadFile("./templates/" + name + ".html")
		if err != nil {
			return err
		}
		tmpls[name] = hold.Parse(name, data)
		return nil
	}
	err := loadTmpl("menu_item")
	if err != nil {
		return err
	}
	err = loadTmpl("menu_alerts")
	if err != nil {
		return err
	}

	var addVariation = func(index int, callback func(mitem MenuItem) bool) {
		renderBuffer, variableIndices := hold.Scan(tmpls, callback)
		hold.Variations[index] = menuTmpl{renderBuffer, variableIndices}
		fmt.Print("renderBuffer: ")
		menuDumpSlice(renderBuffer)
		fmt.Printf("\nvariableIndices: %+v\n", variableIndices)
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
		if tmplData[j] != expects[expectIndex] {
			return j, false
		}
		expectIndex++
	}
	return j, true
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
	//fmt.Println("tmplData: ", string(tmplData))
	var textBuffer, variableBuffer [][]byte
	var renderList []menuRenderItem
	var subBuffer []byte

	// ? We only support simple properties on MenuItem right now
	var addVariable = func(name []byte) {
		//fmt.Println("appending subBuffer: ", string(subBuffer))
		textBuffer = append(textBuffer, subBuffer)
		subBuffer = nil

		//fmt.Println("adding variable: ", string(name))
		variableBuffer = append(variableBuffer, name)
		renderList = append(renderList, menuRenderItem{0, len(textBuffer) - 1})
		renderList = append(renderList, menuRenderItem{1, len(variableBuffer) - 1})
	}

	for i := 0; i < len(tmplData); i++ {
		char := tmplData[i]
		if char == '{' && nextCharIs(tmplData, i, '{') {
			dotIndex, hasDot := skipUntilIfExists(tmplData, i, '.')
			if !hasDot {
				// Template function style
				langIndex, hasChars := skipUntilCharsExist(tmplData, i+2, []byte("lang"))
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
			if !hasFence || !nextCharIs(tmplData, fenceIndex, '}') {
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

	fmt.Println("name: ", name)
	fmt.Print("textBuffer: ")
	menuDumpSlice(textBuffer)
	fmt.Print("\nvariableBuffer: ")
	menuDumpSlice(variableBuffer)
	fmt.Printf("\nrenderList: %+v\n", renderList)
	return MenuTmpl{name, textBuffer, variableBuffer, renderList}
}

func (hold *MenuListHolder) Scan(menuTmpls map[string]MenuTmpl, showItem func(mitem MenuItem) bool) (renderBuffer [][]byte, variableIndices []int) {
	for _, mitem := range hold.List {
		// Do we want this item in this variation of the menu?
		if !showItem(mitem) {
			continue
		}

		menuTmpl, ok := menuTmpls[mitem.TmplName]
		if !ok {
			menuTmpl = menuTmpls["menu_item"]
		}
		fmt.Println("menuTmpl: ", menuTmpl)
		for _, renderItem := range menuTmpl.RenderList {
			if renderItem.Type == 0 {
				renderBuffer = append(renderBuffer, menuTmpl.TextBuffer[renderItem.Index])
				continue
			}

			variable := menuTmpl.VariableBuffer[renderItem.Index]
			fmt.Println("initial variable: ", string(variable))
			dotAt, hasDot := skipUntilIfExists(variable, 0, '.')
			if !hasDot {
				fmt.Println("no dot")
				continue
			}

			if bytes.Equal(variable[:dotAt], []byte("lang")) {
				fmt.Println("lang: ", string(bytes.TrimPrefix(variable[dotAt:], []byte("."))))
				renderBuffer = append(renderBuffer, []byte(GetTmplPhrase(string(bytes.TrimPrefix(variable[dotAt:], []byte("."))))))
			} else {
				var renderItem []byte
				switch string(variable) {
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
						fmt.Println("lang var: ", string(renderItem[dotAt+1:endFence]))
						renderBuffer = append(renderBuffer, []byte(GetTmplPhrase(string(renderItem[dotAt+1:endFence]))))
					} else {
						fmt.Println("other var: ", string(variable[:dotAt]))
						if len(renderItem) > 0 {
							renderBuffer = append(renderBuffer, renderItem)
							variableIndices = append(variableIndices, len(renderBuffer)-1)
						}
					}
					continue
				}

				fmt.Println("normal var: ", string(variable[:dotAt]))
				if len(renderItem) > 0 {
					renderBuffer = append(renderBuffer, renderItem)
				}
			}
		}
	}
	// TODO: Need more coalescing in the renderBuffer
	return renderBuffer, variableIndices
}

// TODO: Pre-render the lang stuff
func (hold *MenuListHolder) Build(w io.Writer, user *User) error {
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

	if len(mTmpl.VariableIndices) == 0 {
		fmt.Println("no variable indices")
		for _, renderItem := range mTmpl.RenderBuffer {
			fmt.Printf("renderItem: %+v\n", renderItem)
			w.Write(renderItem)
		}
		return nil
	}

	var nearIndex = 0
	for index, renderItem := range mTmpl.RenderBuffer {
		if index != mTmpl.VariableIndices[nearIndex] {
			fmt.Println("wrote text: ", string(renderItem))
			w.Write(renderItem)
			continue
		}

		fmt.Println("variable: ", string(renderItem))
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
			if bytes.Equal(variable[fenceStart:dotAt], []byte("me")) {
				fmt.Println("maybe me variable")
				w.Write(variable[prevIndex:fenceStart])
				switch string(variable[dotAt:fenceEnd]) {
				case "Link":
					w.Write([]byte(user.Link))
				case "Session":
					w.Write([]byte(user.Session))
				}
				prevIndex = fenceEnd
			}
		}
		fmt.Println("prevIndex: ", prevIndex)
		fmt.Println("len(variable)-1: ", len(variable)-1)
		w.Write(variable[prevIndex : len(variable)-1])
		if len(mTmpl.VariableIndices) > (nearIndex + 1) {
			nearIndex++
		}
	}
	return nil
}
