package main
import "log"
import "strings"
import "strconv"
import "net/http"
import "golang.org/x/crypto/bcrypt"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

type User struct
{
	ID int
	Name string
	Group int
	Is_Admin bool
	Is_Super_Admin bool
	Is_Banned bool
	Session string
	Loggedin bool
	Avatar string
}

func SetPassword(uid int, password string) (error) {
	salt, err := GenerateSafeString(saltLength)
	if err != nil {
		return err
	}
	
	password = password + salt
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	_, err = set_password_stmt.Exec(string(hashed_password), salt, uid)
	if err != nil {
		return err
	}
	return nil
}

func SessionCheck(w http.ResponseWriter, r *http.Request) (User) {
	user := User{0,"",0,false,false,false,"",false,""}
	var err error
	var cookie *http.Cookie
	
	// Are there any session cookies..?
	// Assign it to user.name to avoid having to create a temporary variable for the type conversion
	cookie, err = r.Cookie("uid")
	if err != nil {
		return user
	}
	user.Name = cookie.Value
	user.ID, err = strconv.Atoi(user.Name)
	if err != nil {
		return user
	}
	cookie, err = r.Cookie("session")
	if err != nil {
		return user
	}
	user.Session = cookie.Value
	//log.Print("ID: " + user.Name)
	//log.Print("Session: " + user.Session)
	
	// Is this session valid..?
	err = get_session_stmt.QueryRow(user.ID,user.Session).Scan(&user.ID, &user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Avatar)
	if err == sql.ErrNoRows {
		return user
	} else if err != nil {
		log.Print(err)
		return user
	}
	user.Is_Admin = (user.Is_Super_Admin || groups[user.Group].Is_Admin)
	user.Is_Banned = groups[user.Group].Is_Banned
	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	user.Loggedin = true
	/*log.Print("Logged in")
	log.Print("ID: " + strconv.Itoa(user.ID))
	log.Print("Group: " + strconv.Itoa(user.Group))
	log.Print("Name: " + user.Name)
	if user.Loggedin {
		log.Print("Loggedin: true")
	} else {
		log.Print("Loggedin: false")
	}
	if user.Is_Admin {
		log.Print("Is_Admin: true")
	} else {
		log.Print("Is_Admin: false")
	}*/
	return user
}