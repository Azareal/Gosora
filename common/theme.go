/* Copyright Azareal 2016 - 2019 */
package common

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	htmpl "html/template"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"strconv"
	"text/template"

	"github.com/Azareal/Gosora/common/phrases"
)

var ErrNoDefaultTheme = errors.New("The default theme isn't registered in the system")
var ErrBadDefaultTemplate = errors.New("The template you tried to load doesn't exist in the interpreted pool.")

type Theme struct {
	Path string // Redirect this file to another folder

	Name           string
	FriendlyName   string
	Version        string
	Creator        string
	FullImage      string
	MobileFriendly bool
	Disabled       bool
	HideFromThemes bool
	BgAvatars      bool // For profiles, at the moment
	GridLists      bool // User Manager
	ForkOf         string
	Tag            string
	URL            string
	Docks          []string // Allowed Values: leftSidebar, rightSidebar, footer
	Settings       map[string]ThemeSetting
	IntTmplHandle  *htmpl.Template
	// TODO: Do we really need both OverridenTemplates AND OverridenMap?
	OverridenTemplates []string
	OverridenMap       map[string]bool
	Templates          []TemplateMapping
	TemplatesMap       map[string]string
	TmplPtr            map[string]interface{}
	Resources          []ThemeResource
	ResourceTemplates  *template.Template

	// Dock intercepters
	// TODO: Implement this
	MapTmplToDock map[string]ThemeMapTmplToDock // map[dockName]data
	RunOnDock     func(string) string           //(dock string) (sbody string)

	// This variable should only be set and unset by the system, not the theme meta file
	// TODO: Should we phase out Active and make the default theme store the primary source of truth?
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
	Async    bool
}

type ThemeMapTmplToDock struct {
	//Name string
	File string
}

// TODO: It might be unsafe to call the template parsing functions with fsnotify, do something more concurrent
func (theme *Theme) LoadStaticFiles() error {
	theme.ResourceTemplates = template.New("")
	fmap := make(map[string]interface{})
	fmap["lang"] = func(phraseNameInt interface{}, tmplInt interface{}) interface{} {
		phraseName, ok := phraseNameInt.(string)
		if !ok {
			panic("phraseNameInt is not a string")
		}
		tmpl, ok := tmplInt.(CSSData)
		if !ok {
			panic("tmplInt is not a CSSData")
		}
		phrase, ok := tmpl.Phrases[phraseName]
		if !ok {
			// TODO: XSS? Only server admins should have access to theme files anyway, but think about it
			return "{lang." + phraseName + "}"
		}
		return phrase
	}
	fmap["toArr"] = func(args ...interface{}) []interface{} {
		return args
	}
	fmap["concat"] = func(args ...interface{}) interface{} {
		var out string
		for _, arg := range args {
			out += arg.(string)
		}
		return out
	}
	theme.ResourceTemplates.Funcs(fmap)
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
			// TODO: Prepare resource templates for each loaded langpack?
			err = theme.ResourceTemplates.ExecuteTemplate(&b, filename, CSSData{Phrases: phraseMap})
			if err != nil {
				log.Print("Failed in adding static file '" + path + "' for default theme '" + theme.Name + "'")
				return err
			}
			data = b.Bytes()
		}

		path = strings.TrimPrefix(path, "themes/"+theme.Name+"/public")
		gzipData, err := CompressBytesGzip(data)
		if err != nil {
			return err
		}

		// Get a checksum for CSPs and cache busting
		hasher := sha256.New()
		hasher.Write(data)
		checksum := hex.EncodeToString(hasher.Sum(nil))

		StaticFiles.Set("/static/"+theme.Name+path, SFile{data, gzipData, checksum,theme.Name+path + "?h=" + checksum, 0, int64(len(data)), int64(len(gzipData)),strconv.Itoa(len(gzipData)), mime.TypeByExtension(ext), f, f.ModTime().UTC().Format(http.TimeFormat)})

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

			dTmplPtr, ok := destTmplPtr.(*func(interface{}, io.Writer) error)
			if !ok {
				log.Print("themeTmpl.Name: ", themeTmpl.Name)
				log.Print("themeTmpl.Source: ", themeTmpl.Source)
				LogError(errors.New("Unknown destination template type!"))
				return
			}

			sTmplPtr, ok := sourceTmplPtr.(*func(interface{}, io.Writer) error)
			if !ok {
				LogError(errors.New("The source and destination templates are incompatible"))
				return
			}

			overridenTemplates[themeTmpl.Name] = true
			*dTmplPtr = *sTmplPtr
		}
	}
}

func (theme *Theme) setActive(active bool) error {
	var sink bool
	err := themeStmts.isDefault.QueryRow(theme.Name).Scan(&sink)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	hasTheme := err != sql.ErrNoRows
	if hasTheme {
		_, err = themeStmts.update.Exec(active, theme.Name)
	} else {
		_, err = themeStmts.add.Exec(theme.Name, active)
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

type GzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w GzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// NEW method of doing theme templates to allow one user to have a different theme to another. Under construction.
// TODO: Generate the type switch instead of writing it by hand
// TODO: Cut the number of types in half
func (theme *Theme) RunTmpl(template string, pi interface{}, w io.Writer) error {
	// Unpack this to avoid an indirect call
	gzw, ok := w.(GzipResponseWriter)
	if ok {
		w = gzw.Writer
	}

	var getTmpl = theme.GetTmpl(template)
	switch tmplO := getTmpl.(type) {
	case *func(interface{}, io.Writer) error:
		var tmpl = *tmplO
		return tmpl(pi, w)
	case func(interface{}, io.Writer) error:
		return tmplO(pi, w)
	case nil, string:
		//fmt.Println("falling back to interpreted for " + template)
		mapping, ok := theme.TemplatesMap[template]
		if !ok {
			mapping = template
		}
		if theme.IntTmplHandle.Lookup(mapping+".html") == nil {
			return ErrBadDefaultTemplate
		}
		return theme.IntTmplHandle.ExecuteTemplate(w, mapping+".html", pi)
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

// GetTmpl attempts to get the template for a specific theme, otherwise it falls back on the default template pointer, which if absent will fallback onto the template interpreter
func (theme *Theme) GetTmpl(template string) interface{} {
	// TODO: Figure out why we're getting a nil pointer here when transpiled templates are disabled, I would have assumed that we would just fall back to !ok on this
	// Might have something to do with it being the theme's TmplPtr map, investigate.
	tmpl, ok := theme.TmplPtr[template]
	if ok {
		return tmpl
	}
	tmpl, ok = TmplPtrMap[template+"_"+theme.Name]
	if ok {
		return tmpl
	}
	tmpl, ok = TmplPtrMap[template]
	if ok {
		return tmpl
	}
	return template
}
