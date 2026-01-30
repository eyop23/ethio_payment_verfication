package models

import (
"go.mongodb.org/mongo-driver/bson/primitive"
"time"
)


// User represents an app user
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Email    string             `bson:"email" json:"email"`
	Phone    string             `bson:"phone" json:"phone"`
	Password string             `bson:"password" json:"-"` // hashed password
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
}
