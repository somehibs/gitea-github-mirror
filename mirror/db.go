package mirror

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Gitea SQLite types from DESCRIBE `table`
type User struct {
	Id               int64
	LowerName        string
	Name             string
	FullName         string
	Email            string
	Type             int
	Location         string
	Website          string
	CreatedUnix      int64
	UpdatedUnix      int64
	LastLoginUnix    int64
	MaxRepoCreation  int
	IsActive         byte
	IsAdmin          byte
	AllowGitHook     byte
	AllowImportLocal byte
	ProhibitLogin    byte
	Avatar           string
	AvatarEmail      string
	NumFollowers     int
	NumFollowing     int
	NumStars         int
	NumRepos         int
	Description      string
	NumTeams         int
	NumMembers       int
	DiffViewStyle    string
}

type Repository struct {
	Id                  int64
	OwnerId             int64
	LowerName           string
	Name                string
	Description         string
	Website             string
	DefaultBranch       string
	NumWatches          int64
	NumStars            int64
	NumForks            int64
	NumIssues           int64
	NumClosedIssues     int64
	NumMilestones       int64
	NumClosedMilestones int64
	IsPrivate           byte
	IsBare              byte
	IsMirror            byte
	IsFork              byte
	ForkId              int64
	Size                int64
	IsFsckEnabled       byte
	Topics              string
	CreatedUnix         int64
	UpdatedUnix         int64
}

type Database struct {
	db *gorm.DB
}

func NewDatabase() Database {
	db := Database{}
	conf := GetConfig()
	var err error
	connectStr := fmt.Sprintf("%s:%s@%s/%s?charset=utf8&parseTime=True", conf.DbUser, conf.DbPass, conf.DbUrl, conf.DbName)
	fmt.Println(connectStr)
	db.db, err = gorm.Open("mysql", connectStr)
	db.db.SingularTable(true)
	if err != nil {
		panic("Cannot open DB: " + err.Error())
	}
	return db
}

func (db *Database) HookUser(username string) {
	// Fetch all of this user's owned repositories by first fetching the user.
	var user User
	db.db.First(&user, "name = ?", username)
	fmt.Printf("Found user for hook: %+v\n", user)
	//userRepos map[string]int
	return
}
