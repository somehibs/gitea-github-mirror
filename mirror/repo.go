package mirror

import (
	"fmt"
	"time"
)

// Repos responsible for
// - receiving webhook updates and dispatching update events to remote endpoints
// - discovering when repositories need new webhooks (intermittent action)
// - determining if there is a sink for a given webhook update

type Repos struct {
	// Keep a copy of the user config
	mirrorUsers map[string]RemoteUser
	// database for webhook queries
	db *Database
}

func NewRepos() Repos {
	r := Repos{}
	r.mirrorUsers = GetUserConfig(GetConfig())
	// start hook job
	db := NewDatabase()
	r.db = &db
	r.checkHooks(&db)
	go r.autohook(60) // run every n seconds
	return r
}

func (r *Repos) checkHooks(db *Database) {
	for username, configUser := range r.mirrorUsers {
		fmt.Printf("User %s and config %+v\n", username, configUser)
		r.enforceUserHooked(db, username, configUser.Ignores)
	}
}

func (r *Repos) enforceUserHooked(db *Database, username string, ignores []string) {
	// Get the count of hooks available
	user := db.User(username)
	if user.Name != username {
		fmt.Printf("Could not find gitea user %s (db gave %s)\n", username, user.Name)
		return
	}
	// Get all the repos for this user (default to no private repos)
	repos := db.UserRepos(user, false)
	hooks := db.RepoHooks(repos)

	hookToRepo := map[int64]Webhook{}
	for _, hook := range hooks {
		hookToRepo[hook.RepoId] = hook
	}

	for _, repo := range repos {
		die := false
		for _, ignore := range ignores {
			if repo.Name == ignore {
				die = true
				break
			}
		}
		if die {
			continue
		}
		if hookToRepo[repo.Id].Id == 0 {
			fmt.Printf("hook does not exist for %s (id: %d)\n", repo.Name, repo.Id)
		}
	}
}

func (r *Repos) autohook(seconds int) {
	db := NewDatabase()
	for true {
		fmt.Printf("d: sleeping for %ds\n", seconds)
		time.Sleep(time.Duration(seconds) * time.Second)
		r.checkHooks(&db)
	}
}
