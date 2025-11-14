package mongoclient

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper functions for building queries and updates

// M is an alias for bson.M for convenience
type M = bson.M

// D is an alias for bson.D for convenience
type D = bson.D

// A is an alias for bson.A for convenience
type A = bson.A

// NewObjectID generates a new MongoDB ObjectID
func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

// ObjectIDFromHex creates an ObjectID from a hex string
func ObjectIDFromHex(hex string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(hex)
}

// IsValidObjectID checks if a string is a valid ObjectID
func IsValidObjectID(hex string) bool {
	_, err := primitive.ObjectIDFromHex(hex)
	return err == nil
}

// Query builders

// Filter creates a new filter (alias for bson.M)
func Filter() M {
	return M{}
}

// Update creates an update document
func Update() M {
	return M{}
}

// Set creates a $set update operator
func Set(fields M) M {
	return M{"$set": fields}
}

// Unset creates a $unset update operator
func Unset(fields ...string) M {
	unsetFields := M{}
	for _, field := range fields {
		unsetFields[field] = ""
	}
	return M{"$unset": unsetFields}
}

// Inc creates a $inc update operator
func Inc(fields M) M {
	return M{"$inc": fields}
}

// Push creates a $push update operator
func Push(field string, value interface{}) M {
	return M{"$push": M{field: value}}
}

// Pull creates a $pull update operator
func Pull(field string, value interface{}) M {
	return M{"$pull": M{field: value}}
}

// AddToSet creates a $addToSet update operator
func AddToSet(field string, value interface{}) M {
	return M{"$addToSet": M{field: value}}
}

// CurrentDate creates a $currentDate update operator
func CurrentDate(fields ...string) M {
	dateFields := M{}
	for _, field := range fields {
		dateFields[field] = true
	}
	return M{"$currentDate": dateFields}
}

// Aggregation pipeline builders

// Pipeline creates a new aggregation pipeline
func Pipeline() A {
	return A{}
}

// Match creates a $match stage
func Match(filter M) M {
	return M{"$match": filter}
}

// Group creates a $group stage
func Group(id interface{}, fields M) M {
	groupFields := M{"_id": id}
	for k, v := range fields {
		groupFields[k] = v
	}
	return M{"$group": groupFields}
}

// Sort creates a $sort stage
func Sort(fields M) M {
	return M{"$sort": fields}
}

// Limit creates a $limit stage
func Limit(n int64) M {
	return M{"$limit": n}
}

// Skip creates a $skip stage
func Skip(n int64) M {
	return M{"$skip": n}
}

// Project creates a $project stage
func Project(fields M) M {
	return M{"$project": fields}
}

// Lookup creates a $lookup stage for joins
func Lookup(from, localField, foreignField, as string) M {
	return M{
		"$lookup": M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	}
}

// Unwind creates a $unwind stage
func Unwind(field string) M {
	return M{"$unwind": field}
}

// Common query operators

// Eq creates an equality condition
func Eq(value interface{}) M {
	return M{"$eq": value}
}

// Ne creates a not equal condition
func Ne(value interface{}) M {
	return M{"$ne": value}
}

// Gt creates a greater than condition
func Gt(value interface{}) M {
	return M{"$gt": value}
}

// Gte creates a greater than or equal condition
func Gte(value interface{}) M {
	return M{"$gte": value}
}

// Lt creates a less than condition
func Lt(value interface{}) M {
	return M{"$lt": value}
}

// Lte creates a less than or equal condition
func Lte(value interface{}) M {
	return M{"$lte": value}
}

// In creates an $in condition
func In(values ...interface{}) M {
	return M{"$in": values}
}

// Nin creates a $nin condition
func Nin(values ...interface{}) M {
	return M{"$nin": values}
}

// Exists creates an $exists condition
func Exists(exists bool) M {
	return M{"$exists": exists}
}

// Regex creates a $regex condition
func Regex(pattern, options string) M {
	return M{"$regex": pattern, "$options": options}
}

// And creates an $and condition
func And(conditions ...M) M {
	return M{"$and": conditions}
}

// Or creates an $or condition
func Or(conditions ...M) M {
	return M{"$or": conditions}
}

// Not creates a $not condition
func Not(condition M) M {
	return M{"$not": condition}
}

// Timestamp helpers

// TimeToTimestamp converts time.Time to MongoDB timestamp
func TimeToTimestamp(t time.Time) primitive.Timestamp {
	return primitive.Timestamp{T: uint32(t.Unix()), I: 0}
}

// TimestampToTime converts MongoDB timestamp to time.Time
func TimestampToTime(ts primitive.Timestamp) time.Time {
	return time.Unix(int64(ts.T), 0)
}

// Pagination helper

// PaginationOptions holds pagination parameters
type PaginationOptions struct {
	Page     int64 // Current page (1-indexed)
	PageSize int64 // Items per page
}

// GetSkip calculates the skip value for pagination
func (p *PaginationOptions) GetSkip() int64 {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size
func (p *PaginationOptions) GetLimit() int64 {
	return p.PageSize
}

// Validate validates pagination options
func (p *PaginationOptions) Validate() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100 // Max page size
	}
}
