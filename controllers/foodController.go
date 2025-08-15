package controllers

import (
	"context"
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
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
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

		matchStage := bson.D{{"%match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"_id", "null"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0},
					{"total_count", 1},
				}
			}
		}
	}
}

func GetFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
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
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
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

	}
}

func round(num float64) int {

}

func toFixed(num float64, precision int) float64 {
	
}