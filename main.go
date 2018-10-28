package main

import (
	"git.circuitco.de/self/gitea-github-mirror/mirror"
)

func main() {
	cfg := mirror.GetConfig()
	_ = mirror.GetUserConfig(cfg)
	mirror.NewDatabase()
	mirror.NewRepos()
	mirror.ListenForever()
}
