/* Copyright Azareal 2016 - 2018 */
package common

import (
	//"fmt"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"

	"../query_gen/lib"
)

type ThemeList map[string]*Theme

var Themes ThemeList = make(map[string]*Theme)
var DefaultThemeBox atomic.Value
var ChangeDefaultThemeMutex sync.Mutex

// TODO: Use this when the default theme doesn't exist
var fallbackTheme = "cosora"
var overridenTemplates = make(map[string]bool)

type Theme struct {
	Name              string
	FriendlyName      string
	Version           string
	Creator           string
	FullImage         string
	MobileFriendly    bool
	Disabled          bool
	HideFromThemes    bool
	BgAvatars         bool // For profiles, at the moment
	ForkOf            string
	Tag               string
	URL               string
	Docks             []string // Allowed Values: leftSidebar, rightSidebar, footer
	Settings          map[string]ThemeSetting
	Templates         []TemplateMapping
	TemplatesMap      map[string]string
	TmplPtr           map[string]interface{}
	Resources         []ThemeResource
	ResourceTemplates *template.Template

	// This variable should only be set and unset by the system, not the theme meta file
	Active bool
}

type ThemeSetting struct {
	FriendlyName string
	Options      []string
}

type TemplateMapping struct {
	Name   string
	Source string
	//When string
}

type ThemeResource struct {
	Name     string
	Location string
}

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

func InitThemes() error {
	themeFiles, err := ioutil.ReadDir("./themes")
	if err != nil {
		return err
	}

	for _, themeFile := range themeFiles {
		if !themeFile.IsDir() {
			continue
		}

		themeName := themeFile.Name()
		log.Printf("Adding theme '%s'", themeName)
		themeFile, err := ioutil.ReadFile("./themes/" + themeName + "/theme.json")
		if err != nil {
			return err
		}

		var theme = &Theme{Name: ""}
		err = json.Unmarshal(themeFile, theme)
		if err != nil {
			return err
		}

		theme.Active = false // Set this to false, just in case someone explicitly overrode this value in the JSON file

		// TODO: Let the theme specify where it's resources are via the JSON file?
		// TODO: Let the theme inherit CSS from another theme?
		// ? - This might not be too helpful, as it only searches for /public/ and not if /public/ is empty. Still, it might help some people with a slightly less cryptic error
		_, err = os.Stat("./themes/" + theme.Name + "/public/")
		if err != nil {
			if os.IsNotExist(err) {
				return errors.New("We couldn't find this theme's resources. E.g. the /public/ folder.")
			} else {
				log.Print("We weren't able to access this theme's resources due to a permissions issue or some other problem")
				return err
			}
		}

		if theme.FullImage != "" {
			DebugLog("Adding theme image")
			err = StaticFiles.Add("./themes/"+themeName+"/"+theme.FullImage, "./themes/"+themeName)
			if err != nil {
				return err
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

		Themes[theme.Name] = theme
	}
	return nil
}

// TODO: It might be unsafe to call the template parsing functions with fsnotify, do something more concurrent
func (theme *Theme) LoadStaticFiles() error {
	theme.ResourceTemplates = template.New("")
	template.Must(theme.ResourceTemplates.ParseGlob("./themes/" + theme.Name + "/public/*.css"))

	// It should be safe for us to load the files for all the themes in memory, as-long as the admin hasn't setup a ridiculous number of themes
	return theme.AddThemeStaticFiles()
}

func (theme *Theme) AddThemeStaticFiles() error {
	phraseMap := GetCSSPhrases()
	// TODO: Use a function instead of a closure to make this more testable? What about a function call inside the closure to take the theme variable into account?
	return filepath.Walk("./themes/"+theme.Name+"/public", func(path string, f os.FileInfo, err error) error {
		DebugLog("Attempting to add static file '" + path + "' for default theme '" + theme.Name + "'")
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		path = strings.Replace(path, "\\", "/", -1)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var ext = filepath.Ext(path)
		if ext == ".css" && len(data) != 0 {
			var b bytes.Buffer
			var pieces = strings.Split(path, "/")
			var filename = pieces[len(pieces)-1]
			err = theme.ResourceTemplates.ExecuteTemplate(&b, filename, CSSData{Phrases: phraseMap})
			if err != nil {
				return err
			}
			data = b.Bytes()
		}

		path = strings.TrimPrefix(path, "themes/"+theme.Name+"/public")
		gzipData := compressBytesGzip(data)
		StaticFiles.Set("/static/"+theme.Name+path, SFile{data, gzipData, 0, int64(len(data)), int64(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)})

		DebugLog("Added the '/" + theme.Name + path + "' static file for theme " + theme.Name + ".")
		return nil
	})
}

func (theme *Theme) MapTemplates() {
	if theme.Templates != nil {
		for _, themeTmpl := range theme.Templates {
			if themeTmpl.Name == "" {
				LogError(errors.New("Invalid destination template name"))
			}
			if themeTmpl.Source == "" {
				LogError(errors.New("Invalid source template name"))
			}

			// `go generate` is one possibility for letting plugins inject custom page structs, but it would simply add another step of compilation. It might be simpler than the current build process from the perspective of the administrator?

			destTmplPtr, ok := TmplPtrMap[themeTmpl.Name]
			if !ok {
				return
			}
			sourceTmplPtr, ok := TmplPtrMap[themeTmpl.Source]
			if !ok {
				LogError(errors.New("The source template doesn't exist!"))
			}

			switch dTmplPtr := destTmplPtr.(type) {
			case *func(TopicPage, http.ResponseWriter) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(TopicPage, http.ResponseWriter) error:
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(TopicsPage, http.ResponseWriter) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(TopicsPage, http.ResponseWriter) error:
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(ForumPage, http.ResponseWriter) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ForumPage, http.ResponseWriter) error:
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(ForumsPage, http.ResponseWriter) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ForumsPage, http.ResponseWriter) error:
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(ProfilePage, http.ResponseWriter) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ProfilePage, http.ResponseWriter) error:
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(CreateTopicPage, http.ResponseWriter) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(CreateTopicPage, http.ResponseWriter) error:
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(IPSearchPage, http.ResponseWriter) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(IPSearchPage, http.ResponseWriter) error:
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(Page, http.ResponseWriter) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(Page, http.ResponseWriter) error:
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			default:
				log.Print("themeTmpl.Name: ", themeTmpl.Name)
				log.Print("themeTmpl.Source: ", themeTmpl.Source)
				LogError(errors.New("Unknown destination template type!"))
			}
		}
	}
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
		case func(TopicPage, http.ResponseWriter) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(TopicPage, http.ResponseWriter) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(TopicsPage, http.ResponseWriter) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(TopicsPage, http.ResponseWriter) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(ForumPage, http.ResponseWriter) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(ForumPage, http.ResponseWriter) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(ForumsPage, http.ResponseWriter) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(ForumsPage, http.ResponseWriter) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(ProfilePage, http.ResponseWriter) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(ProfilePage, http.ResponseWriter) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(CreateTopicPage, http.ResponseWriter) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(CreateTopicPage, http.ResponseWriter) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(IPSearchPage, http.ResponseWriter) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(IPSearchPage, http.ResponseWriter) error:
				*dPtr = oPtr
			default:
				LogError(errors.New("The source and destination templates are incompatible"))
			}
		case func(Page, http.ResponseWriter) error:
			switch dPtr := destTmplPtr.(type) {
			case *func(Page, http.ResponseWriter) error:
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
func RunThemeTemplate(theme string, template string, pi interface{}, w http.ResponseWriter) error {
	var getTmpl = GetThemeTemplate(theme, template)
	switch tmplO := getTmpl.(type) {
	case *func(TopicPage, http.ResponseWriter) error:
		var tmpl = *tmplO
		return tmpl(pi.(TopicPage), w)
	case *func(TopicsPage, http.ResponseWriter) error:
		var tmpl = *tmplO
		return tmpl(pi.(TopicsPage), w)
	case *func(ForumPage, http.ResponseWriter) error:
		var tmpl = *tmplO
		return tmpl(pi.(ForumPage), w)
	case *func(ForumsPage, http.ResponseWriter) error:
		var tmpl = *tmplO
		return tmpl(pi.(ForumsPage), w)
	case *func(ProfilePage, http.ResponseWriter) error:
		var tmpl = *tmplO
		return tmpl(pi.(ProfilePage), w)
	case *func(CreateTopicPage, http.ResponseWriter) error:
		var tmpl = *tmplO
		return tmpl(pi.(CreateTopicPage), w)
	case *func(IPSearchPage, http.ResponseWriter) error:
		var tmpl = *tmplO
		return tmpl(pi.(IPSearchPage), w)
	case *func(Page, http.ResponseWriter) error:
		var tmpl = *tmplO
		return tmpl(pi.(Page), w)
	case func(TopicPage, http.ResponseWriter) error:
		return tmplO(pi.(TopicPage), w)
	case func(TopicsPage, http.ResponseWriter) error:
		return tmplO(pi.(TopicsPage), w)
	case func(ForumPage, http.ResponseWriter) error:
		return tmplO(pi.(ForumPage), w)
	case func(ForumsPage, http.ResponseWriter) error:
		return tmplO(pi.(ForumsPage), w)
	case func(ProfilePage, http.ResponseWriter) error:
		return tmplO(pi.(ProfilePage), w)
	case func(CreateTopicPage, http.ResponseWriter) error:
		return tmplO(pi.(CreateTopicPage), w)
	case func(IPSearchPage, http.ResponseWriter) error:
		return tmplO(pi.(IPSearchPage), w)
	case func(Page, http.ResponseWriter) error:
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

func (theme Theme) HasDock(name string) bool {
	for _, dock := range theme.Docks {
		if dock == name {
			return true
		}
	}
	return false
}

// TODO: Implement this
func (theme Theme) BuildDock(dock string) (sbody string) {
	return ""
}
