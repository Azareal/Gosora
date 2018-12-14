package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Azareal/Gosora/query_gen"
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
		themeStmts = ThemeStmts{
			getAll:    acc.Select("themes").Columns("uname, default").Prepare(),
			isDefault: acc.Select("themes").Columns("default").Where("uname = ?").Prepare(),
			update:    acc.Update("themes").Set("default = ?").Where("uname = ?").Prepare(),
			add:       acc.Insert("themes").Columns("uname, default").Fields("?,?").Prepare(),
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

		var theme = &Theme{Name: ""}
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
func (themes ThemeList) LoadActiveStatus() error {
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
		theme, ok := themes[uname]
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

		themes[uname] = theme
	}
	return rows.Err()
}

func (themes ThemeList) LoadStaticFiles() error {
	for _, theme := range themes {
		err := theme.LoadStaticFiles()
		if err != nil {
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
		switch oPtr := originPointer.(type) {
		case func(CustomPagePage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(CustomPagePage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(TopicPage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(TopicPage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(TopicListPage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(TopicListPage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(ForumPage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(ForumPage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(ForumsPage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(ForumsPage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(ProfilePage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(ProfilePage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(CreateTopicPage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(CreateTopicPage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(IPSearchPage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(IPSearchPage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(AccountDashPage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(AccountDashPage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(ErrorPage, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(ErrorPage, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(Page, io.Writer) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(Page, io.Writer) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		default:
			log.Print("name: ", name)
			LogError(errors.New("Unknown destination template type!"))
		}
		log.Print("The template override was reset")
	}
	overridenTemplates = make(map[string]bool)
	log.Print("All of the template overrides have been reset")
}

// CreateThemeTemplate creates a theme template on the current default theme
func CreateThemeTemplate(theme string, name string) {
	Themes[theme].TmplPtr[name] = func(pi Page, w http.ResponseWriter) error {
		mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap[name]
		if !ok {
			mapping = name
		}
		return Templates.ExecuteTemplate(w, mapping+".html", pi)
	}
}

func GetDefaultThemeName() string {
	return DefaultThemeBox.Load().(string)
}

func SetDefaultThemeName(name string) {
	DefaultThemeBox.Store(name)
}
