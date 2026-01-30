// models/payment.go
package models

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Payment struct {
    ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID                 primitive.ObjectID `bson:"user_id" json:"user_id"`
    TotalAmount            string             `bson:"total_amount" json:"total_amount"`
    PaymentMode            string             `bson:"payment_mode" json:"payment_mode"`
    PaymentReason          string             `bson:"payment_reason" json:"payment_reason"`
    PayerName              string             `bson:"payer_name" json:"payer_name"`
    PayerAccountNo         string             `bson:"payer_account_no" json:"payer_account_no"`
    PaymentDate            time.Time          `bson:"payment_date" json:"payment_date"`
    InvoiceNo              string             `bson:"invoice_no" json:"invoice_no"`
    Status                 string             `bson:"status" json:"status"`
    ReceiverName           string             `bson:"receiver_name" json:"receiver_name"`
    ReceiverAccountNo      string             `bson:"receiver_account_no" json:"receiver_account_no"`
    CreatedAt              time.Time          `bson:"created_at" json:"created_at"`
}