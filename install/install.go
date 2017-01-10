/* Copyright Azareal 2017 - 2018 */
package main

import "fmt"
import "os"
import "bytes"
import "bufio"
import "strconv"
import "io/ioutil"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

var scanner *bufio.Scanner
var db_host string
var db_username string
var db_password string
var db_name string
//var db_collation string = "utf8mb4_general_ci"
var db_port string = "3306"
var site_name string
var site_url string
var server_port string

var default_host string = "localhost"
var default_username string = "root"
var default_dbname string = "gosora"
var default_site_name string = "Site Name"
var default_site_url string = "localhost"
var default_server_port string = "80" // 8080's a good one, if you're testing and don't want it to clash with port 80

func main() {
	scanner = bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to Gosora's Installer")
	fmt.Println("We're going to take you through a few steps to help you get started :)")
	if !get_database_details() {
		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Something went wrong!")
		}
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	if !get_site_details() {
		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Something went wrong!")
		}
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	_db_password := db_password
	if(_db_password != ""){
		_db_password = ":" + _db_password
	}
	db, err := sql.Open("mysql",db_username + _db_password + "@tcp(" + db_host + ":" + db_port + ")/")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	// Make sure that the connection is alive..
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	fmt.Println("Successfully connected to the database")
	fmt.Println("Opening the database seed file")
	sqlContents, err := ioutil.ReadFile("./data.sql")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	var waste string
	err = db.QueryRow("SHOW DATABASES LIKE '" + db_name + "'").Scan(&waste)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	if err == sql.ErrNoRows {
		fmt.Println("Unable to find the database. Attempting to create it")
		_,err = db.Exec("CREATE DATABASE IF NOT EXISTS " + db_name + "")
		if err != nil {
			fmt.Println(err)
			fmt.Println("Aborting installation...")
			press_any_key()
			return
		}
		fmt.Println("The database was successfully created")
	}
	
	fmt.Println("Switching to database " + db_name)
	_, err = db.Exec("USE " + db_name)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	fmt.Println("Preparing installation queries")
	sqlContents = bytes.TrimSpace(sqlContents)
	statements := bytes.Split(sqlContents, []byte(";"))
	for key, statement := range statements {
		if len(statement) == 0 {
			continue
		}
		
		fmt.Println("Executing query #" + strconv.Itoa(key) + " " + string(statement))
		_, err = db.Exec(string(statement))
		if err != nil {
			fmt.Println(err)
			fmt.Println("Aborting installation...")
			press_any_key()
			return
		}
	}
	fmt.Println("Finished inserting the database data")
	
	configContents := []byte(`package main

// Database details
var dbhost = "` + db_host + `"
var dbuser = "` + db_username + `"
var dbpassword = "` + db_password + `"
var dbname = "` + db_name + `"
var dbport = "` + db_port + `" // You probably won't need to change this

// Limiters
var max_request_size = 5 * megabyte

// Misc
var default_route = route_topics
var default_group = 3 // Should be a setting
var activation_group = 5 // Should be a setting
var staff_css = " background-color: #ffeaff;"
var uncategorised_forum_visible = true
var enable_emails = false
var site_name = "` + site_name + `" // Should be a setting
var site_email = "" // Should be a setting
var smtp_server = ""
//var noavatar = "https://api.adorable.io/avatars/{width}/{id}@{site_url}.png"
var noavatar = "https://api.adorable.io/avatars/285/{id}@" + site_url + ".png"
var items_per_page = 40 // Should be a setting

var site_url = "` + site_url + `"
var server_port = "` + server_port + `"
var enable_ssl = false
var ssl_privkey = ""
var ssl_fullchain = ""

// Developer flag
var debug = false
`)
	
	fmt.Println("Opening the configuration file")
	configFile, err := os.Create("./config.go")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	fmt.Println("Writing to the configuration file...")
	_, err = configFile.Write(configContents)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	configFile.Sync()
	configFile.Close()
	fmt.Println("Finished writing to the configuration file")
	
	fmt.Println("Yay, you have successfully installed Gosora!")
	fmt.Println("Your name is Admin and you can login with the password 'password'. Don't forget to change it! Seriously. It's really insecure.")
	press_any_key()
}

func get_database_details() bool {
	fmt.Println("Database Host? Default: " + default_host)
	if !scanner.Scan() {
		return false
	}
	db_host = scanner.Text()
	if db_host == "" {
		db_host = default_host
	}
	fmt.Println("Set database host to " + db_host)
	
	fmt.Println("Database Username? Default: " + default_username)
	if !scanner.Scan() {
		return false
	}
	db_username = scanner.Text()
	if db_username == "" {
		db_username = default_username
	}
	fmt.Println("Set database username to " + db_username)
	
	fmt.Println("Database Password? Default: ''")
	if !scanner.Scan() {
		return false
	}
	db_password = scanner.Text()
	if len(db_password) == 0 {
		fmt.Println("You didn't set a password for this user. This won't block the installation process, but it might create security issues in the future.\n")
	} else {
		fmt.Println("Set password to " + obfuscate_password(db_password))
	}
	
	fmt.Println("Database Name? Pick a name you like or one provided to you. Default: " + default_dbname)
	if !scanner.Scan() {
		return false
	}
	db_name = scanner.Text()
	if db_name == "" {
		db_name = default_dbname
	}
	fmt.Println("Set database name to " + db_name)
	return true
}

func get_site_details() bool {
	fmt.Println("Okay. We also need to know some actual information about your site!")
	fmt.Println("What's your site's name? Default: " + default_site_name)
	if !scanner.Scan() {
		return false
	}
	site_name = scanner.Text()
	if site_name == "" {
		site_name = default_site_name
	}
	fmt.Println("Set the site name to " + site_name)
	
	fmt.Println("What's your site's url? Default: " + default_site_url)
	if !scanner.Scan() {
		return false
	}
	site_url = scanner.Text()
	if site_url == "" {
		site_url = default_site_url
	}
	fmt.Println("Set the site url to " + site_url)
	
	fmt.Println("What port do you want the server to listen on? If you don't know what this means, you should probably leave it on the default. Default: " + default_server_port)
	if !scanner.Scan() {
		return false
	}
	server_port = scanner.Text()
	if server_port == "" {
		server_port = default_server_port
	}
	fmt.Println("Set the server port to " + server_port)
	return true
}

func obfuscate_password(password string) (out string) {
	for i := 0; i < len(password); i++ {
		out += "*"
	}
	return out
}

func press_any_key() {
	//fmt.Println("Press any key to exit...")
	fmt.Println("Please press enter to exit...")
	for scanner.Scan() {
		_ = scanner.Text()
		return
	}
}