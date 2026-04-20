package protocol

// RunRequest is sent from shim to daemon.
type RunRequest struct {
	Argv []string `json:"argv"`
	Cwd  string   `json:"cwd"`
	Env  []string `json:"env"`
}

// RunResponse is returned by the daemon.
type RunResponse struct {
	Allow    bool              `json:"allow"`
	MergeEnv map[string]string `json:"merge_env,omitempty"`
	Message  string            `json:"message,omitempty"`
}
