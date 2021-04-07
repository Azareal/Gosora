package install

import (
	"fmt"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var adapters = make(map[string]InstallAdapter)

type InstallAdapter interface {
	Name() string
	DefaultPort() string
	SetConfig(dbHost, dbUsername, dbPassword, dbName, dbPort string)
	InitDatabase() error
	TableDefs() error
	InitialData() error
	CreateAdmin() error

	DBHost() string
	DBUsername() string
	DBPassword() string
	DBName() string
	DBPort() string
}

func Lookup(name string) (InstallAdapter, bool) {
	adap, ok := adapters[name]
	return adap, ok
}

func createAdmin() error {
	fmt.Println("Creating the admin user")
	hashedPassword, salt, e := BcryptGeneratePassword("password")
	if e != nil {
		return e
	}

	// Build the admin user query
	adminUserStmt, e := qgen.Builder.SimpleInsert("users", "name, password, salt, email, group, is_super_admin, active, createdAt, lastActiveAt, lastLiked, oldestItemLikedCreatedAt, message, last_ip", "'Admin',?,?,'admin@localhost',1,1,1,UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP(),'',''")
	if e != nil {
		return e
	}

	// Run the admin user query
	_, e = adminUserStmt.Exec(hashedPassword, salt)
	return e
}
