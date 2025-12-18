package profile

type GetInput struct {
	UserID string
}
type UpdateInput struct {
	UserID      string
	FirstName   *string
	LastName    *string
	MiddleName  *string
	DisplayName *string
	AvatarURL   *string
}
type Output struct {
	UserID      string
	FirstName   string
	LastName    string
	MiddleName  string
	DisplayName string
	AvatarURL   string
}
