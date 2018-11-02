package mirror

import (
	"fmt"
	"net/http"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

// Repos responsible for
// - receiving webhook updates and dispatching update events to remote endpoints
// - discovering when repositories need new webhooks (intermittent action)
// - determining if there is a sink for a given webhook update

type ServiceHandle func(user, pass, repo, desc string) error

type Repos struct {
	// Keep a copy of the user config
	repoConfig UserConfig
	// database for webhook queries
	db             *Database
	createHandlers map[string]ServiceHandle
}

func NewRepos() Repos {
	r := Repos{}
	r.createHandlers = map[string]ServiceHandle{
		"github": GHCreate,
	}
	r.repoConfig = GetUserConfig(GetConfig())
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

func (r Repos) OpenRepo(name, owner string) *git.Repository {
	root := GetConfig().Path
	path := fmt.Sprintf("%s/%s/%s", root, owner, name)
	path = path + ".git"
	repo, err := git.PlainOpen(path)
	if err != nil {
		fmt.Println("Warning: Failed to open %s for repo name %s with error %s", path, name, err.Error())
	}
	return repo
}

func (r Repos) AddRemote(repo *git.Repository, repoName string, service ServiceConfig, user RemoteUser) {
	url := service.Url
	url = fmt.Sprintf(url, user.Username, user.Token, user.Username+"/"+repoName)
	// first arg is output from git
	_, err := repo.CreateRemote(&gitconfig.RemoteConfig{
		Name: user.Service + "+" + user.Username,
		URLs: []string{url},
	})
	if err != nil && err != git.ErrRemoteExists {
		panic(err.Error())
	}
}

func (r Repos) PushRemote(repo *git.Repository, service ServiceConfig, user RemoteUser, remote, repoName, localUser string, retry bool) {
	err := repo.Push(&git.PushOptions{RemoteName: remote})
	if err == transport.ErrRepositoryNotFound && retry == false {
		if create, ok := r.createHandlers[user.Service]; ok {
			fmt.Println("Attempting to create repository")
			desc := fmt.Sprintf("Automatically mirrored from %s%s/%s", GetConfig().BaseUrl, localUser, repoName)
			create(user.Username, user.Token, repoName, desc)
			r.PushRemote(repo, service, user, remote, repoName, localUser, true)
		} else {
			fmt.Printf("Could not create repository for service type %s so repo %s is being ignored\n", user.Service, repoName)
		}
	} else if err != nil {
		fmt.Printf("Push: Something went wrong (repo %s): %s\n", repoName, err.Error())
	} else {
		fmt.Printf("Pushed %s to %s OK\n", repoName, remote)
	}
}

func (r Repos) Event(req GiteaEvent) {
	owner := req.Repository.Owner.Login
	if localUser, ok := r.repoConfig.Users[owner]; ok {
		// This user is configured.
		repo := r.OpenRepo(req.Repository.Name, owner)
		for _, remoteUser := range localUser.RemoteUsers {
			// Look up each remote config in the remote user config
			if remote, ok := r.repoConfig.RemoteUsers[remoteUser]; ok {
				if service, ok := r.repoConfig.Services[remote.Service]; ok {
					r.AddRemote(repo, req.Repository.Name, service, remote)
					r.PushRemote(repo, service, remote, remoteUser, req.Repository.Name, owner, false)
				}
			} else {
				fmt.Printf("Could not find config for remote %s\n", remoteUser)
			}
		}
	} else {
		fmt.Printf("Push from unknown owner: %+v\n", owner)
		return
	}
	// valid repo, mirror
	// first, get the repo on disk
	//for _, remote := range r.repoConfig[owner] {
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
	for username, configUser := range r.repoConfig.Users {
		//fmt.Printf("Checking hooks for %s with config %+v\n", username, configUser)
		r.enforceUserHooked(db, username, configUser.Ignores)
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
