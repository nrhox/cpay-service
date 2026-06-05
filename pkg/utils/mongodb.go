package utils

import (
	"context"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PaginationParam struct {
	Page      int64
	Limit     int64
	SortBy    string
	SortOrder string
	Filter    bson.M
}

type QueryParams struct {
	Page          int64
	Limit         int64
	SortBy        string
	SortOrder     string
	SearchKeyword string
}

type PaginationResult[T any] struct {
	Data      []T   `json:"data"`
	TotalPage int64 `json:"total_page"`
	TotalData int64 `json:"total_data"`
}

func Paginate[T any](ctx context.Context, collection *mongo.Collection, params PaginationParam) (PaginationResult[T], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Filter == nil {
		params.Filter = bson.M{}
	}

	totalData, err := collection.CountDocuments(ctx, params.Filter)
	if err != nil {
		return PaginationResult[T]{}, err
	}

	skip := (params.Page - 1) * params.Limit
	findOptions := options.Find().SetLimit(params.Limit).SetSkip(skip)

	if params.SortBy != "" {
		order := 1
		if strings.ToLower(params.SortOrder) == "desc" {
			order = -1
		}
		findOptions.SetSort(bson.D{{Key: params.SortBy, Value: order}})
	} else {
		findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	}

	cursor, err := collection.Find(ctx, params.Filter, findOptions)
	if err != nil {
		return PaginationResult[T]{}, err
	}
	defer cursor.Close(ctx)

	var data []T
	if err := cursor.All(ctx, &data); err != nil {
		return PaginationResult[T]{}, err
	}

	if data == nil {
		data = []T{}
	}

	totalPage := totalData / params.Limit
	if totalData%params.Limit != 0 {
		totalPage++
	}

	return PaginationResult[T]{
		Data:      data,
		TotalPage: totalPage,
		TotalData: totalData,
	}, nil
}

func ValidateSortField(allowedFields []string, inputField string, defaultField string) string {
	if slices.Contains(allowedFields, inputField) {
		return inputField
	}
	return defaultField
}
