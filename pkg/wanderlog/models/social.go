package models

// SetLikeRequest represents a request to like/unlike a trip
type SetLikeRequest struct {
	Liked bool `json:"liked"`
}

// LikeCount represents the like status and count for a trip
type LikeCount struct {
	Count     int  `json:"count"`
	UserLiked bool `json:"userLiked"`
	// UserLikedKnown distinguishes a confirmed false value from an unavailable
	// authenticated like state. UserLiked remains for source compatibility.
	UserLikedKnown bool `json:"userLikedKnown"`
}
