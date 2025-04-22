package model

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserTableDB struct {
	ID           string    `json:"id,omitempty"`
	Username     string    `json:"username,omitempty"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"password_hash,omitempty"`
	PasswordSalt string    `json:"password_salt,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	Image        string    `json:"image,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserAuthToken struct {
	ID        string    `json:"id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
	EndDate   time.Time `json:"endDate"`
	UID       string    `json:"uid"`
}
type UserResponse struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Bio       string    `json:"bio"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Token     string    `json:"token"`
}
type UserResponseJSON struct {
	User UserResponse `json:"user"`
}
type Profile struct {
	Id        string    `json:"id"`
	Username  string    `json:"username"`
	Bio       string    `json:"bio"`
	Image     string    `json:"image"`
	Following bool      `json:"following"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ProfileResponse struct {
	Profile Profile `json:"profile"`
}
