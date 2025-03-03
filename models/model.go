package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type BookDetails struct {
	BookID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	BookName     string             `json:"name"`
	Writer       string             `json:"writer"`
	Availability bool               `json:"availability"`
}
