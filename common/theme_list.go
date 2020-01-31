package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	qgen "github.com/Azareal/Gosora/query_gen"
)

// TODO: Something more thread-safe
type ThemeList map[string]*Theme

var Themes ThemeList = make(map[string]*Theme) // ? Refactor this into a store?
var DefaultThemeBox atomic.Value
var ChangeDefaultThemeMutex sync.Mutex

// TODO: Fallback to a random theme if this doesn't exist, so admins can remove themes they don't use
// TODO: Use this when the default theme doesn't exist
var fallbackTheme = "cosora"
var overridenTemplates = make(map[string]bool) // ? What is this used for?

type ThemeStmts struct {
	getAll    *sql.Stmt
	isDefault *sql.Stmt
	update    *sql.Stmt
	add       *sql.Stmt
}

var themeStmts ThemeStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		t := "themes"
		themeStmts = ThemeStmts{
			getAll:    acc.Select(t).Columns("uname,default").Prepare(),
			isDefault: acc.Select(t).Columns("default").Where("uname=?").Prepare(),
			update:    acc.Update(t).Set("default=?").Where("uname=?").Prepare(),
			add:       acc.Insert(t).Columns("uname,default").Fields("?,?").Prepare(),
		}
		return acc.FirstError()
	})
}

func NewThemeList() (themes ThemeList, err error) {
	themes = make(map[string]*Theme)
	themeFiles, err := ioutil.ReadDir("./themes")
	if err != nil {
		return themes, err
	}
	if len(themeFiles) == 0 {
		return themes, errors.New("You don't have any themes")
	}

	var lastTheme, defaultTheme string
	for _, themeFile := range themeFiles {
		if !themeFile.IsDir() {
			continue
		}

		themeName := themeFile.Name()
		log.Printf("Adding theme '%s'", themeName)
		themePath := "./themes/" + themeName
		themeFile, err := ioutil.ReadFile(themePath + "/theme.json")
		if err != nil {
			return themes, err
		}

		theme := &Theme{Name: ""}
		err = json.Unmarshal(themeFile, theme)
		if err != nil {
			return themes, err
		}

		if theme.Name == "" {
			return themes, errors.New("Theme " + themePath + " doesn't have a name set in theme.json")
		}
		if theme.Name == fallbackTheme {
			defaultTheme = fallbackTheme
		}
		lastTheme = theme.Name

		// TODO: Implement the static file part of this and fsnotify
		if theme.Path != "" {
			log.Print("Resolving redirect to " + theme.Path)
			themeFile, err := ioutil.ReadFile(theme.Path + "/theme.json")
			if err != nil {
				return themes, err
			}
			theme = &Theme{Name: "", Path: theme.Path}
			err = json.Unmarshal(themeFile, theme)
			if err != nil {
				return themes, err
			}
		} else {
			theme.Path = themePath
		}

		theme.Active = false // Set this to false, just in case someone explicitly overrode this value in the JSON file

		// TODO: Let the theme specify where it's resources are via the JSON file?
		// TODO: Let the theme inherit CSS from another theme?
		// ? - This might not be too helpful, as it only searches for /public/ and not if /public/ is empty. Still, it might help some people with a slightly less cryptic error
		log.Print(theme.Path + "/public/")
		_, err = os.Stat(theme.Path + "/public/")
		if err != nil {
			if os.IsNotExist(err) {
				return themes, errors.New("We couldn't find this theme's resources. E.g. the /public/ folder.")
			} else {
				log.Print("We weren't able to access this theme's resources due to a permissions issue or some other problem")
				return themes, err
			}
		}

		if theme.FullImage != "" {
			DebugLog("Adding theme image")
			err = StaticFiles.Add(theme.Path+"/"+theme.FullImage, themePath)
			if err != nil {
				return themes, err
			}
		}

		theme.TemplatesMap = make(map[string]string)
		theme.TmplPtr = make(map[string]interface{})
		if theme.Templates != nil {
			for _, themeTmpl := range theme.Templates {
				theme.TemplatesMap[themeTmpl.Name] = themeTmpl.Source
				theme.TmplPtr[themeTmpl.Name] = TmplPtrMap["o_"+themeTmpl.Source]
			}
		}

		theme.IntTmplHandle = DefaultTemplates
		overrides, err := ioutil.ReadDir(theme.Path + "/overrides/")
		if err != nil && !os.IsNotExist(err) {
			return themes, err
		}
		if len(overrides) > 0 {
			overCount := 0
			theme.OverridenMap = make(map[string]bool)
			for _, override := range overrides {
				if override.IsDir() {
					continue
				}
				ext := filepath.Ext(themePath + "/overrides/" + override.Name())
				log.Print("attempting to add " + themePath + "/overrides/" + override.Name())
				if ext != ".html" {
					log.Print("not a html file")
					continue
				}
				overCount++
				nosuf := strings.TrimSuffix(override.Name(), ext)
				theme.OverridenTemplates = append(theme.OverridenTemplates, nosuf)
				theme.OverridenMap[nosuf] = true
				//theme.TmplPtr[nosuf] = TmplPtrMap["o_"+nosuf]
				log.Print("succeeded")
			}

			localTmpls := template.New("")
			err = loadTemplates(localTmpls, theme.Name)
			if err != nil {
				return themes, err
			}
			theme.IntTmplHandle = localTmpls
			log.Printf("theme.OverridenTemplates: %+v\n", theme.OverridenTemplates)
			log.Printf("theme.IntTmplHandle: %+v\n", theme.IntTmplHandle)
		} else {
			log.Print("no overrides for " + theme.Name)
		}

		// TODO: Bind the built template, or an interpreted one for any dock overrides this theme has

		themes[theme.Name] = theme
	}
	if defaultTheme == "" {
		defaultTheme = lastTheme
	}
	DefaultThemeBox.Store(defaultTheme)

	return themes, nil
}

// TODO: Make the initThemes and LoadThemes functions less confusing
// ? - Delete themes which no longer exist in the themes folder from the database?
func (t ThemeList) LoadActiveStatus() error {
	ChangeDefaultThemeMutex.Lock()
	defer ChangeDefaultThemeMutex.Unlock()

	rows, err := themeStmts.getAll.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var uname string
	var defaultThemeSwitch bool
	for rows.Next() {
		err = rows.Scan(&uname, &defaultThemeSwitch)
		if err != nil {
			return err
		}

		// Was the theme deleted at some point?
		theme, ok := t[uname]
		if !ok {
			continue
		}

		if defaultThemeSwitch {
			DebugLogf("Loading the default theme '%s'", theme.Name)
			theme.Active = true
			DefaultThemeBox.Store(theme.Name)
			theme.MapTemplates()
		} else {
			DebugLogf("Loading the theme '%s'", theme.Name)
			theme.Active = false
		}

		t[uname] = theme
	}
	return rows.Err()
}

func (t ThemeList) LoadStaticFiles() error {
	for _, theme := range t {
		if err := theme.LoadStaticFiles(); err != nil {
			return err
		}
	}
	return nil
}

func ResetTemplateOverrides() {
	log.Print("Resetting the template overrides")
	for name := range overridenTemplates {
		log.Print("Resetting '" + name + "' template override")
		originPointer, ok := TmplPtrMap["o_"+name]
		if !ok {
			log.Print("The origin template doesn't exist!")
			return
		}
		destTmplPtr, ok := TmplPtrMap[name]
		if !ok {
			log.Print("The destination template doesn't exist!")
			return
		}

		// Not really a pointer, more of a function handle, an artifact from one of the earlier versions of themes.go
		oPtr, ok := originPointer.(func(interface{}, io.Writer) error)
		if !ok {
			log.Print("name: ", name)
			LogError(errors.New("Unknown destination template type!"))
			return
		}

		dPtr, ok := destTmplPtr.(*func(interface{}, io.Writer) error)
		if !ok {
			LogError(errors.New("The source and destination templates are incompatible"))
			return
		}
		*dPtr = oPtr
		log.Print("The template override was reset")
	}
	overridenTemplates = make(map[string]bool)
	log.Print("All of the template overrides have been reset")
}

// CreateThemeTemplate creates a theme template on the current default theme
func CreateThemeTemplate(theme, name string) {
	Themes[theme].TmplPtr[name] = func(pi Page, w http.ResponseWriter) error {
		mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap[name]
		if !ok {
			mapping = name
		}
		return DefaultTemplates.ExecuteTemplate(w, mapping+".html", pi)
	}
}

func GetDefaultThemeName() string {
	return DefaultThemeBox.Load().(string)
}

func SetDefaultThemeName(name string) {
	DefaultThemeBox.Store(name)
}
