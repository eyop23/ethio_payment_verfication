package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Provider represents a payment provider like TeleBirr or CBE
type Provider struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name"`
	URL  string             `bson:"url" json:"url"`
}
