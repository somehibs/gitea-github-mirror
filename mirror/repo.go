package mirror

import (
	"fmt"
	"net/http"
	"time"

	git "gopkg.in/src-d/go-git.v4"
)

// Repos responsible for
// - receiving webhook updates and dispatching update events to remote endpoints
// - discovering when repositories need new webhooks (intermittent action)
// - determining if there is a sink for a given webhook update

type Repos struct {
	// Keep a copy of the user config
	mirrorUsers map[string][]RemoteUser
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
	// start webhooks
	whl := NewWebhookListener(r.Event)
	whl.Handle("/status", r.ServeStatus)
	return r
}

func (r Repos) Event(req GiteaEvent) {
	owner := req.Repository.Owner.Login
	if r.mirrorUsers[owner] == nil {
		// Couldn't find owner with valid remote user
		fmt.Printf("Push from unknown owner: %+v\n", owner)
		return
	}
	// valid repo, mirror
	// first, get the repo on disk
	root := GetConfig().Path
	path := fmt.Sprintf("%s/%s/%s", root, owner, req.Repository.Name)
	path = path + ".git"
	repo, err := git.PlainOpen(path)
	fmt.Printf("Path: %s Repo: %+v err %+v\n", path, repo, err)
	//for _, remote := range r.mirrorUsers[owner] {
	//	remote
	//}
}

func (r *Repos) ServeStatus(writer http.ResponseWriter, request *http.Request) {
	// Status request. Report last run, number of active webhooks, number of recognised repos, number of recognised users.
	// Report status of syncing to remote repositories.
	writer.Write([]byte("Hello world"))
	return
}

func (r *Repos) checkHooks(db *Database) {
	for username, configUser := range r.mirrorUsers {
		fmt.Printf("Checking hooks for %s with config %+v\n", username, configUser)
		r.enforceUserHooked(db, username, configUser[0].Ignores)
	}
}

func (r *Repos) IsIgnored(ignores []string, repoName string) bool {
	for _, ignore := range ignores {
		if repoName == ignore {
			return true
		}
	}
	return false
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

	hookToRepo := map[int64][]Webhook{}
	for _, hook := range hooks {
		hookToRepo[hook.RepoId] = append(hookToRepo[hook.RepoId], hook)
	}

	for _, repo := range repos {
		if r.IsIgnored(ignores, repo.Name) {
			continue
		}
		// Check for our metadata in this repo's hooks
		foundMetadata := false
		meta := "Gitea Github Mirror Webhook"
		for _, hook := range hookToRepo[repo.Id] {
			if hook.Meta == meta {
				foundMetadata = true
				break
			}
		}
		if foundMetadata == false || hookToRepo[repo.Id] == nil {
			fmt.Printf("adding hook for %s (id: %d)\n", repo.Name, repo.Id)
			db.AddRepoHook(fmt.Sprintf("http://%s:%d/", GetConfig().IP, GetConfig().Port), "", true, false, meta, repo)
		}
	}
}

func (r *Repos) autohook(seconds int) {
	db := NewDatabase()
	for true {
		time.Sleep(time.Duration(seconds) * time.Second)
		r.checkHooks(&db)
	}
}
