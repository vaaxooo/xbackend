package refresh

type Input struct {
	RefreshToken string
}
type Output struct {
	AccessToken  string
	RefreshToken string
}
