package models

// SendInvitesRequest represents a request to send trip invitations
type SendInvitesRequest struct {
	Invitees []string `json:"invitees"` // Email addresses or user IDs
	Message  string   `json:"message,omitempty"`
}

// TripInvite represents an invitation to collaborate on a trip
type TripInvite struct {
	Email     string `json:"email"`
	InvitedAt string `json:"invitedAt"`
	Status    string `json:"status"` // e.g., "pending", "accepted"
}

// ShareKeyPermissions represents permissions for a share key
type ShareKeyPermissions struct {
	CanEdit bool `json:"canEdit"`
	CanView bool `json:"canView"`
}

// ShareKeyResponse represents the response from creating/getting a share key
type ShareKeyResponse struct {
	ShareKey string `json:"shareKey"`
}

// CollaboratorRequest represents a request to add/remove a collaborator
type CollaboratorRequest struct {
	UserID int `json:"userId"`
}
