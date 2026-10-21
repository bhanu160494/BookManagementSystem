package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type BookDetails struct {
	BookID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	BookName     string             `json:"name"`
	Writer       string             `json:"writer"`
	Availability bool               `json:"availability"`
}

// SignupRequest is the payload accepted by POST /auth/signup.
type SignupRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
}

// SigninRequest is the payload accepted by POST /auth/signin.
type SigninRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest is the payload accepted by POST /auth/refresh and /auth/logout.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}
