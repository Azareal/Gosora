/* Copyright Azareal 2016 - 2019 */
package common

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Azareal/Gosora/common/phrases"
)

var ErrNoDefaultTheme = errors.New("The default theme isn't registered in the system")

type Theme struct {
	Path string // Redirect this file to another folder

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

	// Dock intercepters
	// TODO: Implement this
	MapTmplToDock map[string]ThemeMapTmplToDock // map[dockName]data
	RunOnDock     func(string) string           //(dock string) (sbody string)

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
	Loggedin bool // Only serve this resource to logged in users
}

type ThemeMapTmplToDock struct {
	//Name string
	File string
}

// TODO: It might be unsafe to call the template parsing functions with fsnotify, do something more concurrent
func (theme *Theme) LoadStaticFiles() error {
	theme.ResourceTemplates = template.New("")
	template.Must(theme.ResourceTemplates.ParseGlob("./themes/" + theme.Name + "/public/*.css"))

	// It should be safe for us to load the files for all the themes in memory, as-long as the admin hasn't setup a ridiculous number of themes
	return theme.AddThemeStaticFiles()
}

func (theme *Theme) AddThemeStaticFiles() error {
	phraseMap := phrases.GetTmplPhrases()
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
		gzipData, err := compressBytesGzip(data)
		if err != nil {
			return err
		}

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
			case *func(CustomPagePage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(CustomPagePage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(TopicPage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(TopicPage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(TopicListPage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(TopicListPage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(ForumPage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ForumPage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(ForumsPage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ForumsPage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(ProfilePage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ProfilePage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(CreateTopicPage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(CreateTopicPage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(IPSearchPage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(IPSearchPage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(AccountDashPage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(AccountDashPage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(ErrorPage, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(ErrorPage, io.Writer) error:
					overridenTemplates[themeTmpl.Name] = true
					*dTmplPtr = *sTmplPtr
				default:
					LogError(errors.New("The source and destination templates are incompatible"))
				}
			case *func(Page, io.Writer) error:
				switch sTmplPtr := sourceTmplPtr.(type) {
				case *func(Page, io.Writer) error:
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

func (theme *Theme) setActive(active bool) error {
	var sink bool
	err := themeStmts.isThemeDefault.QueryRow(theme.Name).Scan(&sink)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	hasTheme := err != sql.ErrNoRows
	if hasTheme {
		_, err = themeStmts.updateTheme.Exec(active, theme.Name)
	} else {
		_, err = themeStmts.addTheme.Exec(theme.Name, active)
	}
	if err != nil {
		return err
	}

	// TODO: Think about what we want to do for multi-server configurations
	log.Printf("Setting theme '%s' as the default theme", theme.Name)
	theme.Active = active
	return nil
}

func UpdateDefaultTheme(theme *Theme) error {
	ChangeDefaultThemeMutex.Lock()
	defer ChangeDefaultThemeMutex.Unlock()

	err := theme.setActive(true)
	if err != nil {
		return err
	}

	defaultTheme := DefaultThemeBox.Load().(string)
	dtheme, ok := Themes[defaultTheme]
	if !ok {
		return ErrNoDefaultTheme
	}

	err = dtheme.setActive(false)
	if err != nil {
		return err
	}

	DefaultThemeBox.Store(theme.Name)
	ResetTemplateOverrides()
	theme.MapTemplates()

	return nil
}

func (theme Theme) HasDock(name string) bool {
	for _, dock := range theme.Docks {
		if dock == name {
			return true
		}
	}
	return false
}

func (theme Theme) BuildDock(dock string) (sbody string) {
	runOnDock := theme.RunOnDock
	if runOnDock != nil {
		return runOnDock(dock)
	}
	return ""
}
