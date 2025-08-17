package controllers

import (
	"context"
	"log"
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

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		result, err := orderCollection.Find(c, bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items"})
			return
		}

		var allorders []bson.M
		if err = result.All(c, &allorders); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allorders)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		orderId := ctx.Param("order_id")
		var order models.Order

		err := orderCollection.FindOne(c, bson.M{"order_id": orderId}).Decode(&order)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching order"})
			return
		}
		ctx.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		var order models.Order
		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(order)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		var table models.Table
		if order.Table_id != nil {
			err := tableCollection.FindOne(c, bson.M{"table_id": order.Table_id}).Decode(&table)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "table was not found"})
				return
			}
		}

		order.Created_at = time.Now().UTC()
		order.Updated_at = time.Now().UTC()

		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()

		result, insertErr := orderCollection.InsertOne(c, order)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "order item was not created"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		var table models.Table
		var order models.Order

		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orderId := ctx.Param("order_id")
		var updateObj primitive.D

		if order.Table_id != nil {
			err := menuCollection.FindOne(c, bson.M{"table_id": order.Table_id}).Decode(&table)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "table not found"})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "order_id", Value: order.Order_id})
		}

		order.Updated_at = time.Now().UTC()
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.Updated_at})

		upsert := true
		filter := bson.M{"order_id": orderId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderCollection.UpdateOne(
			c,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "order item update failed"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func OrderItemOrderCreator(order models.Order) string {
	c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	order.Created_at = time.Now().UTC()
	order.Updated_at = time.Now().UTC()

	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(c, order)

	return order.Order_id
}
