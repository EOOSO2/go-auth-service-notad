package entity

type AuthClaims struct {
	UserID       string   `json:"user_id"`
	EmailID      string   `json:"email_id"`
	EmailAddress string   `json:"email"`
	Permission   []string `json:"permission"`
}
