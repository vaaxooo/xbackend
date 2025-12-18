package dto

type UpdateProfileRequest struct {
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	MiddleName  *string `json:"middle_name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type LinkProviderRequest struct {
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id"`
}

type LinkProviderResponse struct {
	Linked bool `json:"linked"`
}
