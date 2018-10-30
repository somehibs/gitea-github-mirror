package mirror

import (
	"encoding/json"
	"fmt"

	"net/http"
)

type Hook func(event GiteaEvent)

type Author struct {
	Name     string
	Email    string
	Username string
}

type GiteaUser struct {
	Id        int64
	Login     string
	FullName  string
	Email     string
	AvatarUrl string
	Language  string
	Username  string
}

type Commit struct {
	Id           string
	Message      string
	Url          string
	Author       Author
	Committer    Author
	Verification interface{}
	Timestamp    string
}

type ApiRepository struct {
	Id              int64
	Owner           GiteaUser
	Name            string
	FullName        string
	Description     string
	Empty           bool
	Private         bool
	Fork            bool
	Parent          interface{}
	Mirror          bool
	Size            int64
	HtmlUrl         string
	SshUrl          string
	CloneUrl        string
	Website         string
	StarsCount      int64
	ForksCount      int64
	WatchersCount   int64
	OpenIssuesCount int64
	DefaultBranch   string
	CreatedAt       string
	UpdatedAt       string
}

type GiteaEvent struct {
	Secret     string
	Ref        string
	Before     string
	After      string
	CompareUrl string
	Commits    []Commit
	Repository ApiRepository
	Pusher     GiteaUser
	Sender     GiteaUser
}

type WebhookListener struct {
	hook Hook
}

var started = false

func NewWebhookListener(hook Hook) WebhookListener {
	wh := WebhookListener{hook}
	http.Handle("/", wh)
	http.Handle("/query", wh)
	return wh
}

func (w WebhookListener) Handle(uri string, listener http.HandlerFunc) {
	if started {
		panic("Adding handler after starting listener")
	}
	http.Handle(uri, listener)
}

func (w WebhookListener) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		decoder := json.NewDecoder(request.Body)
		var push GiteaEvent
		decoder.Decode(&push)
		w.hook(push)
	}
}

func ListenForever() {
	started = true
	http.ListenAndServe(fmt.Sprintf("%s:%d", GetConfig().IP, GetConfig().Port), nil)
}
