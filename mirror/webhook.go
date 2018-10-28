package mirror

import (
	"fmt"

	"net/http"
)

type Hook interface {
	Called(request *Request)
}

type Request struct {
	Url  string
	Repo string
	User string
	Body string
}

type WebhookListener struct {
	// Uses net/http
	hooks []Hook
}

func NewWebhookListener() WebhookListener {
	hook := WebhookListener{}
	http.Handle("/", hook)
	return hook
}

func (w WebhookListener) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// Got a message from gitea
}

func (w *WebhookListener) AddHook(hook Hook) {
	if w.HasHook(hook) {
		return
	}
	w.hooks = append(w.hooks, hook)
}

func (w *WebhookListener) HasHook(hook Hook) bool {
	for _, hk := range w.hooks {
		if hk == hook {
			return true
		}
	}
	return false
}

func ListenForever() {
	http.ListenAndServe(fmt.Sprintf(":%d", GetConfig().Port), nil)
}
