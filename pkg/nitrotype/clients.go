package nitrotype

type APIClient interface {
	GetTeam(tagName string) (*TeamAPIResponse, error)
	GetProfile(username string) (*UserProfile, error)
}
