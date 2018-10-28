package mirror

// Repos responsible for
// - receiving webhook updates and dispatching update events to remote endpoints
// - discovering when repositories need new webhooks (intermittent action)
// - determining if there is a sink for a given webhook update

type Repos struct {
	// Keep a copy of the user config
	config map[string]User
}

func NewRepos() Repos {
	r := Repos{}
	r.config = GetUserConfig(GetConfig())
	return r
}
