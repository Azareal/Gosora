package main
import "strings"
import "os"
import "log"
import "io/ioutil"
import "path/filepath"

type Page struct
{
	Title string
	Name string
	CurrentUser User
	ItemList map[int]interface{}
	Something interface{}
}

func add_custom_page(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	
	// Is this a directory..?
	fileInfo, err := os.Stat(path)
    is_dir := fileInfo.IsDir()
	if err != nil {
		return err
	}
	if is_dir {
		return err
	}
	
	custom_page, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	log.Print("Loaded the '" + path + "' page.")
	name := strings.TrimSuffix(path, filepath.Ext(path))
	custom_pages[name] = string(custom_page)
	return nil
}