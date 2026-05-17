package queries

type CASStatusResp struct {
	Provider      string `json:"provider"`
	Enabled       bool   `json:"enabled"`
	Configured    bool   `json:"configured"`
	AutoProvision bool   `json:"auto_provision"`
	LoginPath     string `json:"login_path"`
	CallbackPath  string `json:"callback_path"`
}

type CASLoginResp struct {
	Provider    string `json:"provider"`
	RedirectURL string `json:"redirect_url"`
	CallbackURL string `json:"callback_url"`
}
