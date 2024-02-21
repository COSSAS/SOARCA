package projections

import "go.mongodb.org/mongo-driver/bson"

type Projection uint8

const (
	Id Projection = iota
	Meta
)

func (p Projection) GetProjection() bson.D {
	switch p {
	case Id:
		return bson.D{{Key: "_id", Value: 1}}
	case Meta:
		return bson.D{
			{Key: "_id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "description", Value: 1},
			{Key: "created", Value: 1},
			{Key: "valid_from", Value: 1},
			{Key: "valid_until", Value: 1},
			{Key: "labels", Value: 1},
		}
	}
	return bson.D{}
}
