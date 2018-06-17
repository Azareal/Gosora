package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"strconv"

	"../common"
	"../query_gen/lib"
	_ "github.com/go-sql-driver/mysql"
)

var patches = make(map[int]func(*bufio.Scanner) error)

func addPatch(index int, handle func(*bufio.Scanner) error) {
	patches[index] = handle
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Capture panics instead of closing the window at a superhuman speed before the user can read the message on Windows
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
			debug.PrintStack()
			pressAnyKey(scanner)
			log.Fatal("")
			return
		}
	}()

	log.Print("Loading the configuration data")
	err := common.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Processing configuration data")
	err = common.ProcessConfig()
	if err != nil {
		log.Fatal(err)
	}

	if common.DbConfig.Adapter != "mysql" && common.DbConfig.Adapter != "" {
		log.Fatal("Only MySQL is supported for upgrades right now, please wait for a newer build of the patcher")
	}

	err = prepMySQL()
	if err != nil {
		log.Fatal(err)
	}

	err = patcher(scanner)
	if err != nil {
		log.Fatal(err)
	}
}

func pressAnyKey(scanner *bufio.Scanner) {
	fmt.Println("Please press enter to exit...")
	for scanner.Scan() {
		_ = scanner.Text()
		return
	}
}

func prepMySQL() error {
	return qgen.Builder.Init("mysql", map[string]string{
		"host":      common.DbConfig.Host,
		"port":      common.DbConfig.Port,
		"name":      common.DbConfig.Dbname,
		"username":  common.DbConfig.Username,
		"password":  common.DbConfig.Password,
		"collation": "utf8mb4_general_ci",
	})
}

type SchemaFile struct {
	DBVersion          string // Current version of the database schema
	DynamicFileVersion string
	MinGoVersion       string // TODO: Minimum version of Go needed to install this version
	MinVersion         string // TODO: Minimum version of Gosora to jump to this version, might be tricky as we don't store this in the schema file, maybe store it in the database
}

func patcher(scanner *bufio.Scanner) error {
	fmt.Println("Loading the schema file")
	data, err := ioutil.ReadFile("./schema/lastSchema.json")
	if err != nil {
		return err
	}

	var schemaFile SchemaFile
	err = json.Unmarshal(data, &schemaFile)
	if err != nil {
		return err
	}
	dbVersion, err := strconv.Atoi(schemaFile.DBVersion)
	if err != nil {
		return err
	}

	fmt.Println("Applying the patches")
	var pslice = make([]func(*bufio.Scanner) error, len(patches))
	for i := 0; i < len(patches); i++ {
		pslice[i] = patches[i]
	}

	// Run the queued up patches
	for index, patch := range pslice {
		if dbVersion > index {
			continue
		}
		err := patch(scanner)
		if err != nil {
			return err
		}
	}

	return nil
}
