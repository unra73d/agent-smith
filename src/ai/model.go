package ai

type Model struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider IAPIProvider
}

func LoadModels() []*Model {
	log.D("Loading models")

	models := make([]*Model, 0, 32)

	providers := LoadProviders()
	for _, provider := range providers {
		providerModels, err := provider.ListModels()
		if err != nil {
			log.E("Failed to list models for provider", provider.Name(), err)
			continue
		}
		models = append(models, providerModels...)
	}
	return models
}
