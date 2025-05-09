package ai

type Model struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider IAPIProvider
}
