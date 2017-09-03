/* Copyright Azareal 2016 - 2018 */
package main

import (
	//"fmt"
	"bytes"
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
	"text/template"
)

var defaultTheme string
var themes = make(map[string]Theme)

//var overridenTemplates map[string]interface{} = make(map[string]interface{})
var overridenTemplates = make(map[string]bool)

type Theme struct {
	Name           string
	FriendlyName   string
	Version        string
	Creator        string
	FullImage      string
	MobileFriendly bool
	Disabled       bool
	HideFromThemes bool
	ForkOf         string
	Tag            string
	URL            string
	Sidebars       string // Allowed Values: left, right, both, false
	//DisableMinifier // Is this really a good idea? I don't think themes should be fighting against the minifier
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

func LoadThemes() error {
	rows, err := get_themes_stmt.Query()
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

		theme.TemplatesMap = make(map[string]string)
		theme.TmplPtr = make(map[string]interface{})
		if theme.Templates != nil {
			for _, themeTmpl := range theme.Templates {
				theme.TemplatesMap[themeTmpl.Name] = themeTmpl.Source
				theme.TmplPtr[themeTmpl.Name] = tmplPtrMap["o_"+themeTmpl.Source]
			}
		}

		theme.ResourceTemplates = template.New("")
		template.Must(theme.ResourceTemplates.ParseGlob("./themes/" + uname + "/public/*.css"))

		if defaultThemeSwitch {
			log.Print("Loading the theme '" + theme.Name + "'")
			theme.Active = true
			defaultTheme = uname
			mapThemeTemplates(theme)
		} else {
			theme.Active = false
		}

		// It should be safe for us to load the files for all the themes in memory, as-long as the admin hasn't setup a ridiculous number of themes
		err = addThemeStaticFiles(theme)
		if err != nil {
			return err
		}
		themes[uname] = theme
	}
	return rows.Err()
}

func initThemes() error {
	themeFiles, err := ioutil.ReadDir("./themes")
	if err != nil {
		return err
	}

	for _, themeFile := range themeFiles {
		if !themeFile.IsDir() {
			continue
		}

		themeName := themeFile.Name()
		log.Print("Adding theme '" + themeName + "'")
		themeFile, err := ioutil.ReadFile("./themes/" + themeName + "/theme.json")
		if err != nil {
			return err
		}

		var theme Theme
		err = json.Unmarshal(themeFile, &theme)
		if err != nil {
			return err
		}

		theme.Active = false // Set this to false, just in case someone explicitly overrode this value in the JSON file

		if theme.FullImage != "" {
			if dev.DebugMode {
				log.Print("Adding theme image")
			}
			err = addStaticFile("./themes/"+themeName+"/"+theme.FullImage, "./themes/"+themeName)
			if err != nil {
				return err
			}
		}

		themes[theme.Name] = theme
	}
	return nil
}

func addThemeStaticFiles(theme Theme) error {
	// TO-DO: Use a function instead of a closure to make this more testable? What about a function call inside the closure to take the theme variable into account?
	return filepath.Walk("./themes/"+theme.Name+"/public", func(path string, f os.FileInfo, err error) error {
		if dev.DebugMode {
			log.Print("Attempting to add static file '" + path + "' for default theme '" + theme.Name + "'")
		}
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
		//log.Print("path ",path)
		//log.Print("ext ",ext)
		if ext == ".css" && len(data) != 0 {
			var b bytes.Buffer
			var pieces = strings.Split(path, "/")
			var filename = pieces[len(pieces)-1]
			//log.Print("filename ", filename)
			err = theme.ResourceTemplates.ExecuteTemplate(&b, filename, CssData{ComingSoon: "We don't have any data to pass you yet!"})
			if err != nil {
				return err
			}
			data = b.Bytes()
		}

		path = strings.TrimPrefix(path, "themes/"+theme.Name+"/public")
		gzipData := compressBytesGzip(data)
		staticFiles["/static/"+theme.Name+path] = SFile{data, gzipData, 0, int64(len(data)), int64(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)}

		if dev.DebugMode {
			log.Print("Added the '/" + theme.Name + path + "' static file for theme " + theme.Name + ".")
		}
		return nil
	})
}

func mapThemeTemplates(theme Theme) {
	if theme.Templates != nil {
		for _, themeTmpl := range theme.Templates {
			if themeTmpl.Name == "" {
				log.Fatal("Invalid destination template name")
			}
			if themeTmpl.Source == "" {
				log.Fatal("Invalid source template name")
			}

			// `go generate` is one possibility for letting plugins inject custom page structs, but it would simply add another step of compilation. It might be simpler than the current build process from the perspective of the administrator?

			destTmplPtr, ok := tmplPtrMap[themeTmpl.Name]
			if !ok {
				return
			}
			sourceTmplPtr, ok := tmplPtrMap[themeTmpl.Source]
			if !ok {
				log.Fatal("The source template doesn't exist!")
			}

			switch dTmplPtr := destTmplPtr.(type) {
			case *func(TopicPage, http.ResponseWriter):
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(TopicPage, http.ResponseWriter):
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					log.Fatal("The source and destination templates are incompatible")
				}
			case *func(TopicsPage, http.ResponseWriter):
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(TopicsPage, http.ResponseWriter):
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					log.Fatal("The source and destination templates are incompatible")
				}
			case *func(ForumPage, http.ResponseWriter):
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ForumPage, http.ResponseWriter):
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					log.Fatal("The source and destination templates are incompatible")
				}
			case *func(ForumsPage, http.ResponseWriter):
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ForumsPage, http.ResponseWriter):
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					log.Fatal("The source and destination templates are incompatible")
				}
			case *func(ProfilePage, http.ResponseWriter):
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ProfilePage, http.ResponseWriter):
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					log.Fatal("The source and destination templates are incompatible")
				}
			case *func(CreateTopicPage, http.ResponseWriter):
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(CreateTopicPage, http.ResponseWriter):
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					log.Fatal("The source and destination templates are incompatible")
				}
			case *func(Page, http.ResponseWriter):
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(Page, http.ResponseWriter):
					//overridenTemplates[themeTmpl.Name] = d_tmpl_ptr
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					log.Fatal("The source and destination templates are incompatible")
				}
			default:
				log.Fatal("Unknown destination template type!")
			}
		}
	}
}

func resetTemplateOverrides() {
	log.Print("Resetting the template overrides")

	for name := range overridenTemplates {
		log.Print("Resetting '" + name + "' template override")

		originPointer, ok := tmplPtrMap["o_"+name]
		if !ok {
			//log.Fatal("The origin template doesn't exist!")
			log.Print("The origin template doesn't exist!")
			return
		}

		destTmplPtr, ok := tmplPtrMap[name]
		if !ok {
			//log.Fatal("The destination template doesn't exist!")
			log.Print("The destination template doesn't exist!")
			return
		}

		// Not really a pointer, more of a function handle, an artifact from one of the earlier versions of themes.go
		switch oPtr := originPointer.(type) {
		case func(TopicPage, http.ResponseWriter):
			switch dPtr := destTmplPtr.(type) {
			case *func(TopicPage, http.ResponseWriter):
				*dPtr = oPtr
			default:
				log.Fatal("The origin and destination templates are incompatible")
			}
		case func(TopicsPage, http.ResponseWriter):
			switch dPtr := destTmplPtr.(type) {
			case *func(TopicsPage, http.ResponseWriter):
				*dPtr = oPtr
			default:
				log.Fatal("The origin and destination templates are incompatible")
			}
		case func(ForumPage, http.ResponseWriter):
			switch dPtr := destTmplPtr.(type) {
			case *func(ForumPage, http.ResponseWriter):
				*dPtr = oPtr
			default:
				log.Fatal("The origin and destination templates are incompatible")
			}
		case func(ForumsPage, http.ResponseWriter):
			switch dPtr := destTmplPtr.(type) {
			case *func(ForumsPage, http.ResponseWriter):
				*dPtr = oPtr
			default:
				log.Fatal("The origin and destination templates are incompatible")
			}
		case func(ProfilePage, http.ResponseWriter):
			switch dPtr := destTmplPtr.(type) {
			case *func(ProfilePage, http.ResponseWriter):
				*dPtr = oPtr
			default:
				log.Fatal("The origin and destination templates are incompatible")
			}
		case func(CreateTopicPage, http.ResponseWriter):
			switch dPtr := destTmplPtr.(type) {
			case *func(CreateTopicPage, http.ResponseWriter):
				*dPtr = oPtr
			default:
				log.Fatal("The origin and destination templates are incompatible")
			}
		case func(Page, http.ResponseWriter):
			switch dPtr := destTmplPtr.(type) {
			case *func(Page, http.ResponseWriter):
				*dPtr = oPtr
			default:
				log.Fatal("The origin and destination templates are incompatible")
			}
		default:
			log.Fatal("Unknown destination template type!")
		}
		log.Print("The template override was reset")
	}
	overridenTemplates = make(map[string]bool)
	log.Print("All of the template overrides have been reset")
}

// NEW method of doing theme templates to allow one user to have a different theme to another. Under construction.
// TO-DO: Generate the type switch instead of writing it by hand
// TO-DO: Cut the number of types in half
func RunThemeTemplate(theme string, template string, pi interface{}, w http.ResponseWriter) {
	switch tmplO := GetThemeTemplate(theme, template).(type) {
	case *func(TopicPage, http.ResponseWriter):
		var tmpl = *tmplO
		tmpl(pi.(TopicPage), w)
	case *func(TopicsPage, http.ResponseWriter):
		var tmpl = *tmplO
		tmpl(pi.(TopicsPage), w)
	case *func(ForumPage, http.ResponseWriter):
		var tmpl = *tmplO
		tmpl(pi.(ForumPage), w)
	case *func(ForumsPage, http.ResponseWriter):
		var tmpl = *tmplO
		tmpl(pi.(ForumsPage), w)
	case *func(ProfilePage, http.ResponseWriter):
		var tmpl = *tmplO
		tmpl(pi.(ProfilePage), w)
	case *func(CreateTopicPage, http.ResponseWriter):
		var tmpl = *tmplO
		tmpl(pi.(CreateTopicPage), w)
	case *func(Page, http.ResponseWriter):
		var tmpl = *tmplO
		tmpl(pi.(Page), w)
	case func(TopicPage, http.ResponseWriter):
		tmplO(pi.(TopicPage), w)
	case func(TopicsPage, http.ResponseWriter):
		tmplO(pi.(TopicsPage), w)
	case func(ForumPage, http.ResponseWriter):
		tmplO(pi.(ForumPage), w)
	case func(ForumsPage, http.ResponseWriter):
		tmplO(pi.(ForumsPage), w)
	case func(ProfilePage, http.ResponseWriter):
		tmplO(pi.(ProfilePage), w)
	case func(CreateTopicPage, http.ResponseWriter):
		tmplO(pi.(CreateTopicPage), w)
	case func(Page, http.ResponseWriter):
		tmplO(pi.(Page), w)
	case string:
		mapping, ok := themes[defaultTheme].TemplatesMap[template]
		if !ok {
			mapping = template
		}
		err := templates.ExecuteTemplate(w, mapping+".html", pi)
		if err != nil {
			LogError(err)
		}
	default:
		log.Print("theme ", theme)
		log.Print("template ", template)
		log.Print("pi ", pi)
		log.Print("tmplO ", tmplO)

		valueOf := reflect.ValueOf(tmplO)
		log.Print("initial valueOf.Type()", valueOf.Type())
		for valueOf.Kind() == reflect.Interface || valueOf.Kind() == reflect.Ptr {
			valueOf = valueOf.Elem()
			log.Print("valueOf.Elem().Type() ", valueOf.Type())
		}
		log.Print("deferenced valueOf.Type() ", valueOf.Type())
		log.Print("valueOf.Kind() ", valueOf.Kind())

		LogError(errors.New("Unknown template type"))
	}
}

// GetThemeTemplate attempts to get the template for a specific theme, otherwise it falls back on the default template pointer, which if absent will fallback onto the template interpreter
func GetThemeTemplate(theme string, template string) interface{} {
	tmpl, ok := themes[theme].TmplPtr[template]
	if ok {
		return tmpl
	}
	tmpl, ok = tmplPtrMap[template]
	if ok {
		return tmpl
	}
	return template
}

// CreateThemeTemplate creates a theme template on the current default theme
func CreateThemeTemplate(theme string, name string) {
	themes[theme].TmplPtr[name] = func(pi Page, w http.ResponseWriter) {
		mapping, ok := themes[defaultTheme].TemplatesMap[name]
		if !ok {
			mapping = name
		}
		err := templates.ExecuteTemplate(w, mapping+".html", pi)
		if err != nil {
			InternalError(err, w)
		}
	}
}
