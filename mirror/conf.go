package mirror

import (
	"fmt"

	"encoding/json"
	"os"

	"github.com/namsral/flag"
)

type Config struct {
	DbUser     string
	DbPass     string
	DbUrl      string
	DbName     string
	Port       int
	UserConfig string
}

var cfg Config

func GetConfig() Config {
	if cfg.DbUser != "" {
		return cfg
	}
	user := flag.String("db_user", "gitea", "Gitea db user (defaults to gitea)")
	pass := flag.String("db_pass", "", "Gitea db password")
	db := flag.String("db", "gitea", "Gitea db (defaults to gitea)")
	url := flag.String("db_url", "127.0.0.1", "Mysql DB url (defaults to 127.0.0.1)")
	port := flag.Int("port", 9001, "Webhook port (defaults to 9001)")
	users := flag.String("user_conf", "user.json", "File containing mappings of gitea users to github credentials")
	fmt.Println("Loading config...")
	flag.Parse()
	cfg = Config{*user, *pass, *url, *db, *port, *users}
	return cfg
}

type RemoteUser struct {
	Username string
	Token    string
	Ignores  []string
}

func GetUserConfig(cfg Config) map[string]RemoteUser {
	f, err := os.Open(cfg.UserConfig)
	if err != nil {
		panic("Error opening user config: " + err.Error())
	}
	defer f.Close()
	decode := json.NewDecoder(f)
	userMap := map[string]RemoteUser{}
	err = decode.Decode(&userMap)
	if err != nil {
		panic("Error reading user config: " + err.Error())
	}
	return userMap
}
