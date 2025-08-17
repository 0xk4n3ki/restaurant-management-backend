package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/0xk4n3ki/restaurant-management-backend/database"
	"github.com/0xk4n3ki/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		result, err := tableCollection.Find(c, bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing table items"})
			return
		}

		var allTables []bson.M
		if err = result.All(c, &allTables); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error while decoding tables"})
			return
		}
		ctx.JSON(http.StatusOK, allTables)
	}
}

func GetTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		tableId := ctx.Param("table_id")
		var table models.Table

		err := tableCollection.FindOne(c, bson.M{"table_id": tableId}).Decode(&table)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
			return
		}
		ctx.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		var table models.Table

		if err := ctx.BindJSON(&table); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(table)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		table.Created_at = time.Now().UTC()
		table.Updated_at = time.Now().UTC()
		table.ID = primitive.NewObjectID()
		table.Table_id = table.ID.Hex()

		result, insertErr := tableCollection.InsertOne(c, table)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create table item"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		var table models.Table
		tableId := ctx.Param("table_id")

		if err := ctx.BindJSON(&table); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if table.Number_of_guests != nil {
			updateObj = append(updateObj, bson.E{Key: "number_of_guests", Value: table.Number_of_guests})
		}
		if table.Table_number != nil {
			updateObj = append(updateObj, bson.E{Key: "table_number", Value: table.Table_number})
		}

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		filter := bson.M{"table_id": tableId}

		result, err := tableCollection.UpdateOne(
			c,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "table item update failed"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
