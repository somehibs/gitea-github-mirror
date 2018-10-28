package mirror

import (
	//"fmt"

	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Gitea SQLite types from DESCRIBE `table`
type Webhook struct {
	Id           int64
	RepoId       int64
	OrgId        int64
	Url          string
	ContentType  int64
	Secret       string
	Events       string
	IsSsl        byte
	IsActive     byte
	HookTaskType int
	Meta         string
	LastStatus   int
	CreatedUnix  int64
	UpdatedUnix  int64
}

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
	db.db, err = gorm.Open("mysql", connectStr)
	db.db.SingularTable(true)
	if err != nil {
		panic("Cannot open DB: " + err.Error())
	}
	return db
}

func (db *Database) User(username string) User {
	var user User
	db.db.First(&user, "name = ?", username)
	return user
}

func (db *Database) UserRepos(user User, private bool) []Repository {
	repos := make([]Repository, 0)
	db.db.Where("owner_id = ? AND is_private = ?", user.Id, private).Find(&repos)
	return repos
}

func (db *Database) RepoHooks(repos []Repository) []Webhook {
	hooks := make([]Webhook, 0)
	repoIds := make([]int64, len(repos))
	for _, repo := range repos {
		repoIds = append(repoIds, repo.Id)
	}
	db.db.Where("repo_id in (?)", repoIds).Find(&hooks)
	return hooks
}
