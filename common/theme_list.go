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
	"reflect"
	"sync"
	"sync/atomic"

	"../query_gen/lib"
)

type ThemeList map[string]*Theme

var Themes ThemeList = make(map[string]*Theme) // ? Refactor this into a store?
var DefaultThemeBox atomic.Value
var ChangeDefaultThemeMutex sync.Mutex

// TODO: Use this when the default theme doesn't exist
var fallbackTheme = "cosora"
var overridenTemplates = make(map[string]bool) // ? What is this used for?

type ThemeStmts struct {
	getThemes *sql.Stmt
}

var themeStmts ThemeStmts

func init() {
	DefaultThemeBox.Store(fallbackTheme)
	DbInits.Add(func(acc *qgen.Accumulator) error {
		themeStmts = ThemeStmts{
			getThemes: acc.Select("themes").Columns("uname, default").Prepare(),
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
	return themes, nil
}

// TODO: Make the initThemes and LoadThemes functions less confusing
// ? - Delete themes which no longer exist in the themes folder from the database?
func (themes ThemeList) LoadActiveStatus() error {
	ChangeDefaultThemeMutex.Lock()
	rows, err := themeStmts.getThemes.Query()
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
			log.Printf("Loading the default theme '%s'", theme.Name)
			theme.Active = true
			DefaultThemeBox.Store(theme.Name)
			theme.MapTemplates()
		} else {
			log.Printf("Loading the theme '%s'", theme.Name)
			theme.Active = false
		}

		themes[uname] = theme
	}
	ChangeDefaultThemeMutex.Unlock()
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

// NEW method of doing theme templates to allow one user to have a different theme to another. Under construction.
// TODO: Generate the type switch instead of writing it by hand
// TODO: Cut the number of types in half
func RunThemeTemplate(theme string, template string, pi interface{}, w io.Writer) error {
	var getTmpl = GetThemeTemplate(theme, template)
	switch tmplO := getTmpl.(type) {
	case *func(TopicPage, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi.(TopicPage), w)
	case *func(TopicListPage, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi.(TopicListPage), w)
	case *func(ForumPage, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi.(ForumPage), w)
	case *func(ForumsPage, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi.(ForumsPage), w)
	case *func(ProfilePage, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi.(ProfilePage), w)
	case *func(CreateTopicPage, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi.(CreateTopicPage), w)
	case *func(IPSearchPage, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi.(IPSearchPage), w)
	case *func(Page, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi.(Page), w)
	case func(TopicPage, io.Writer) error:
		return tmplO(pi.(TopicPage), w)
	case func(TopicListPage, io.Writer) error:
		return tmplO(pi.(TopicListPage), w)
	case func(ForumPage, io.Writer) error:
		return tmplO(pi.(ForumPage), w)
	case func(ForumsPage, io.Writer) error:
		return tmplO(pi.(ForumsPage), w)
	case func(ProfilePage, io.Writer) error:
		return tmplO(pi.(ProfilePage), w)
	case func(CreateTopicPage, io.Writer) error:
		return tmplO(pi.(CreateTopicPage), w)
	case func(IPSearchPage, io.Writer) error:
		return tmplO(pi.(IPSearchPage), w)
	case func(Page, io.Writer) error:
		return tmplO(pi.(Page), w)
	case string:
		mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap[template]
		if !ok {
			mapping = template
		}
		return Templates.ExecuteTemplate(w, mapping+".html", pi)
	default:
		log.Print("theme ", theme)
		log.Print("template ", template)
		log.Print("pi ", pi)
		log.Print("tmplO ", tmplO)
		log.Print("getTmpl ", getTmpl)

		valueOf := reflect.ValueOf(tmplO)
		log.Print("initial valueOf.Type()", valueOf.Type())
		for valueOf.Kind() == reflect.Interface || valueOf.Kind() == reflect.Ptr {
			valueOf = valueOf.Elem()
			log.Print("valueOf.Elem().Type() ", valueOf.Type())
		}
		log.Print("deferenced valueOf.Type() ", valueOf.Type())
		log.Print("valueOf.Kind() ", valueOf.Kind())

		return errors.New("Unknown template type")
	}
}

// GetThemeTemplate attempts to get the template for a specific theme, otherwise it falls back on the default template pointer, which if absent will fallback onto the template interpreter
func GetThemeTemplate(theme string, template string) interface{} {
	tmpl, ok := Themes[theme].TmplPtr[template]
	if ok {
		return tmpl
	}
	tmpl, ok = TmplPtrMap[template]
	if ok {
		return tmpl
	}
	return template
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
