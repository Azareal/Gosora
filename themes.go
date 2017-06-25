/* Copyright Azareal 2016 - 2018 */
package main

import (
	//"fmt"
	"log"
	"io"
	"os"
	"bytes"
	"strings"
	"mime"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
	"net/http"
	"text/template"
)

var defaultTheme string
var themes map[string]Theme = make(map[string]Theme)
//var overriden_templates map[string]interface{} = make(map[string]interface{})
var overriden_templates map[string]bool = make(map[string]bool)

type Theme struct
{
	Name string
	FriendlyName string
	Version string
	Creator string
	FullImage string
	MobileFriendly bool
	Disabled bool
	HideFromThemes bool
	ForkOf string
	Tag string
	URL string
	Sidebars string // Allowed Values: left, right, both, false
	Settings map[string]ThemeSetting
	Templates []TemplateMapping
	TemplatesMap map[string]string
	Resources []ThemeResource
	ResourceTemplates *template.Template

	// This variable should only be set and unset by the system, not the theme meta file
	Active bool
}

type ThemeSetting struct
{
	FriendlyName string
	Options []string
}

type TemplateMapping struct
{
	Name string
	Source string
	//When string
}

type ThemeResource struct
{
	Name string
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
		if theme.Templates != nil {
			for _, themeTmpl := range theme.Templates {
				theme.TemplatesMap[themeTmpl.Name] = themeTmpl.Source
			}
		}

		theme.ResourceTemplates = template.New("")
		template.Must(theme.ResourceTemplates.ParseGlob("./themes/" + uname + "/public/*.css"))

		if defaultThemeSwitch {
			log.Print("Loading the theme '" + theme.Name + "'")
			theme.Active = true
			defaultTheme = uname
			add_theme_static_files(theme)
			map_theme_templates(theme)
		} else {
			theme.Active = false
		}
		themes[uname] = theme
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func init_themes() {
	themeFiles, err := ioutil.ReadDir("./themes")
	if err != nil {
		log.Fatal(err)
	}

	for _, themeFile := range themeFiles {
		if !themeFile.IsDir() {
			continue
		}

		themeName := themeFile.Name()
		log.Print("Adding theme '" + themeName + "'")
		themeFile, err := ioutil.ReadFile("./themes/" + themeName + "/theme.json")
		if err != nil {
			log.Fatal(err)
		}

		var theme Theme
		err = json.Unmarshal(themeFile, &theme)
		if err != nil {
			log.Fatal(err)
		}

		theme.Active = false // Set this to false, just in case someone explicitly overrode this value in the JSON file

		if theme.FullImage != "" {
			if debug {
				log.Print("Adding theme image")
			}
			err = add_static_file("./themes/" + themeName + "/" + theme.FullImage, "./themes/" + themeName)
			if err != nil {
				log.Fatal(err)
			}
		}

		themes[theme.Name] = theme
	}
}

func add_theme_static_files(theme Theme) {
	err := filepath.Walk("./themes/" + theme.Name + "/public", func(path string, f os.FileInfo, err error) error {
		if debug {
			log.Print("Attempting to add static file '" + path + "' for default theme '" + theme.Name + "'")
		}
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		path = strings.Replace(path,"\\","/",-1)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var ext string = filepath.Ext(path)
		//log.Print("path ",path)
		//log.Print("ext ",ext)
		if ext == ".css" {
			var b bytes.Buffer
			var pieces []string = strings.Split(path,"/")
			var filename string = pieces[len(pieces) - 1]
			//log.Print("filename ", filename)
			err = theme.ResourceTemplates.ExecuteTemplate(&b,filename, CssData{ComingSoon:"We don't have any data to pass you yet!"})
			if err != nil {
				return err
			}
			data = b.Bytes()
		}

		path = strings.TrimPrefix(path,"themes/" + theme.Name + "/public")
		gzip_data := compress_bytes_gzip(data)
		static_files["/static" + path] = SFile{data,gzip_data,0,int64(len(data)),int64(len(gzip_data)),mime.TypeByExtension(ext),f,f.ModTime().UTC().Format(http.TimeFormat)}

		if debug {
			log.Print("Added the '" + path + "' static file for default theme " + theme.Name + ".")
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func map_theme_templates(theme Theme) {
	if theme.Templates != nil {
		for _, themeTmpl := range theme.Templates {
			if themeTmpl.Name == "" {
				log.Fatal("Invalid destination template name")
			}
			if themeTmpl.Source == "" {
				log.Fatal("Invalid source template name")
			}

			// `go generate` is one possibility for letting plugins inject custom page structs, but it would simply add another step of compilation. It might be simpler than the current build process from the perspective of the administrator?

			dest_tmpl_ptr, ok := tmpl_ptr_map[themeTmpl.Name]
			if !ok {
				return
			}
			source_tmpl_ptr, ok := tmpl_ptr_map[themeTmpl.Source]
			if !ok {
				log.Fatal("The source template doesn't exist!")
			}

			switch d_tmpl_ptr := dest_tmpl_ptr.(type) {
				case *func(TopicPage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(TopicPage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(TopicsPage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(TopicsPage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(ForumPage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(ForumPage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(ForumsPage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(ForumsPage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(ProfilePage,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(ProfilePage,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				case *func(Page,io.Writer):
					switch s_tmpl_ptr := source_tmpl_ptr.(type) {
						case *func(Page,io.Writer):
							//overriden_templates[themeTmpl.Name] = d_tmpl_ptr
							overriden_templates[themeTmpl.Name] = true
							*d_tmpl_ptr = *s_tmpl_ptr
						default:
							log.Fatal("The source and destination templates are incompatible")
					}
				default:
					log.Fatal("Unknown destination template type!")
			}
		}
	}
}

func reset_template_overrides() {
	log.Print("Resetting the template overrides")

	for name, _ := range overriden_templates {
		log.Print("Resetting '" + name + "' template override")

		origin_pointer, ok := tmpl_ptr_map["o_" + name]
		if !ok {
			//log.Fatal("The origin template doesn't exist!")
			log.Print("The origin template doesn't exist!")
			return
		}

		dest_tmpl_ptr, ok := tmpl_ptr_map[name]
		if !ok {
			//log.Fatal("The destination template doesn't exist!")
			log.Print("The destination template doesn't exist!")
			return
		}

		// Not really a pointer, more of a function handle, an artifact from one of the earlier versions of themes.go
		switch o_ptr := origin_pointer.(type) {
			case func(TopicPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(TopicPage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			case func(TopicsPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(TopicsPage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			case func(ForumPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ForumPage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			case func(ForumsPage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ForumsPage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			case func(ProfilePage,io.Writer):
				switch d_ptr := dest_tmpl_ptr.(type) {
					case *func(ProfilePage,io.Writer):
						*d_ptr = o_ptr
					default:
						log.Fatal("The origin and destination templates are incompatible")
				}
			default:
				log.Fatal("Unknown destination template type!")
		}
		log.Print("The template override was reset")
	}
	overriden_templates = make(map[string]bool)
	log.Print("All of the template overrides have been reset")
}
