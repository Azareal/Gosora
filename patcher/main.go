package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"strconv"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
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
		if r := recover() r != nil {
			fmt.Println(r)
			debug.PrintStack()
			pressAnyKey(scanner)
			log.Fatal("")
			return
		}
	}()

	log.Print("Loading the configuration data")
	err := c.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Processing configuration data")
	err = c.ProcessConfig()
	if err != nil {
		log.Fatal(err)
	}

	if c.DbConfig.Adapter != "mysql" && c.DbConfig.Adapter != "" {
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
		"host":      c.DbConfig.Host,
		"port":      c.DbConfig.Port,
		"name":      c.DbConfig.Dbname,
		"username":  c.DbConfig.Username,
		"password":  c.DbConfig.Password,
		"collation": "utf8mb4_general_ci",
	})
}

type SchemaFile struct {
	DBVersion          string // Current version of the database schema
	DynamicFileVersion string
	MinGoVersion       string // TODO: Minimum version of Go needed to install this version
	MinVersion         string // TODO: Minimum version of Gosora to jump to this version, might be tricky as we don't store this in the schema file, maybe store it in the database
}

func loadSchema() (schemaFile SchemaFile, err error) {
	fmt.Println("Loading the schema file")
	data, err := ioutil.ReadFile("./schema/lastSchema.json")
	if err != nil {
		return schemaFile, err
	}
	err = json.Unmarshal(data, &schemaFile)
	return schemaFile, err
}

func patcher(scanner *bufio.Scanner) error {
	var dbVersion int
	err := qgen.NewAcc().Select("updates").Columns("dbVersion").QueryRow().Scan(&dbVersion)
	if err == sql.ErrNoRows {
		schemaFile, err := loadSchema()
		if err != nil {
			return err
		}
		dbVersion, err = strconv.Atoi(schemaFile.DBVersion)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	fmt.Println("Applying the patches")
	var pslice = make([]func(*bufio.Scanner) error, len(patches))
	for i := 0; i < len(patches); i++ {
		pslice[i] = patches[i]
	}

	// Run the queued up patches
	var patched int
	for index, patch := range pslice {
		if dbVersion > index {
			continue
		}
		err := patch(scanner)
		if err != nil {
			fmt.Println("Failed to apply patch "+strconv.Itoa(index+1))
			return err
		}
		fmt.Println("Applied patch "+strconv.Itoa(index+1))
		patched++
	}

	if patched > 0 {
		_, err := qgen.NewAcc().Update("updates").Set("dbVersion = ?").Exec(len(pslice))
		if err != nil {
			return err
		}
	} else {
		fmt.Println("No new patches found.")
	}

	return nil
}
