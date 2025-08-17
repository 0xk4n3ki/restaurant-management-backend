package controllers

import (
	"context"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/0xk4n3ki/restaurant-management-backend/database"
	"github.com/0xk4n3ki/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex, err := strconv.Atoi(ctx.Query("startIndex"))
		if err != nil || startIndex < 0 {
			startIndex = (page-1)*recordPerPage
		}

		matchStage := bson.D{{Key:"%match", Value:bson.D{{}}}}
		groupStage := bson.D{{
			Key:"$group", 
			Value:bson.D{
				{Key:"_id", Value:bson.D{{Key:"_id", Value:"null"}}}, 
				{Key:"total_count", Value:bson.D{{Key:"$sum", Value:1}}}, 
				{Key:"data", Value:bson.D{{Key:"$push", Value:"$$ROOT"}}},
			}}}
		projectStage := bson.D{{
			Key:"$project", 
			Value:bson.D{
				{Key:"_id", Value:0},
				{Key:"total_count", Value:1},
				{Key:"food_items", Value:bson.D{{Key:"$slice", Value:[]interface{} {"$data", startIndex, recordPerPage}}}},
			}}}
		result, err := foodCollection.Aggregate(c, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing food items"})
			return 
		}

		var allFoods []bson.M
		if err = result.All(c, &allFoods); err != nil {
			log.Fatal(err)
		}

		ctx.JSON(http.StatusOK, allFoods[0])
	}
}

func GetFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		foodId := ctx.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(c, bson.M{"food_id": foodId}).Decode(&food)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while fetching the food item"})
		}

		ctx.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		var menu models.Menu
		var food models.Food

		if err := ctx.BindJSON(&food); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return 
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return 
		}

		err := menuCollection.FindOne(c, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error":"menu was not found"})
			return 
		}

		food.Created_at = time.Now().UTC()
		food.Updated_at = time.Now().UTC()
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		num := toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(c, food)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error":"food item was not created"})
			return 
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 100*time.Second)
		defer cancel()

		var menu models.Menu
		var food models.Food
		foodId := ctx.Param("food_id")

		if err := ctx.BindJSON(&food); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return 
		}

		var updateObj primitive.D
		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}
		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		}
		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Food_image})
		}
		if food.Menu_id != nil {
			err := menuCollection.FindOne(c, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error":"menu was not found"})
				return 
			}
			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: food.Menu_id})
		}
		
		food.Updated_at = time.Now().UTC()
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.Updated_at})

		upsert := true
		filter := bson.M{"food_id": foodId}

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := foodCollection.UpdateOne(
			c,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "food item update failed"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num * output)) / output
}