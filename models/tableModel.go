package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct{
	ID    				primitive.ObjectID			`bson:"_id"`
	Number_of_guests	*int						`json:"number_of_guests" validate:"required"`
	Table_number		*int						`json:"table_number" validate:"required"`
	Updated_at			time.Time					`json:"created_at"`
	Created_at			time.Time					`json:"updated_at"`
	Table_id            string						`json:"table_no"`
}