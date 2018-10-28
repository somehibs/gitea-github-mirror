package mirror

import (
	"fmt"
)

// Repos responsible for
// - receiving webhook updates and dispatching update events to remote endpoints
// - discovering when repositories need new webhooks (intermittent action)
// - determining if there is a sink for a given webhook update

type Repos struct {
	// Keep a copy of the user config
	config map[string]RemoteUser
	// database for webhook queries
	db *Database
}

func NewRepos() Repos {
	r := Repos{}
	r.config = GetUserConfig(GetConfig())
	// start hook job
	db := NewDatabase()
	r.db = &db
	r.checkHooks(&db)
	return r
}

func (r *Repos) checkHooks(db *Database) {
	for username, configUser := range r.config {
		fmt.Printf("User %s and config %+v\n", username, configUser)
		db.HookUser(username)
	}
}

func (r *Repos) autohook() {
	//db := NewDatabase()
	for true {
		break
	}
}
