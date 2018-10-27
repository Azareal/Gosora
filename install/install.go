package install

import (
	"fmt"

	"github.com/Azareal/Gosora/query_gen"
)

var adapters = make(map[string]InstallAdapter)

type InstallAdapter interface {
	Name() string
	DefaultPort() string
	SetConfig(dbHost string, dbUsername string, dbPassword string, dbName string, dbPort string)
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
	hashedPassword, salt, err := BcryptGeneratePassword("password")
	if err != nil {
		return err
	}

	// Build the admin user query
	adminUserStmt, err := qgen.Builder.SimpleInsert("users", "name, password, salt, email, group, is_super_admin, active, createdAt, lastActiveAt, lastLiked, oldestItemLikedCreatedAt, message, last_ip", "'Admin',?,?,'admin@localhost',1,1,1,UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP(),'','127.0.0.1'")
	if err != nil {
		return err
	}

	// Run the admin user query
	_, err = adminUserStmt.Exec(hashedPassword, salt)
	return err
}
