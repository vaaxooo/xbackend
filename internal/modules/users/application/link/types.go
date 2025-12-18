package link

type Input struct {
	UserID         string
	Provider       string
	ProviderUserID string
}

type Output struct {
	Linked bool
}
