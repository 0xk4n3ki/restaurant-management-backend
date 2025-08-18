package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/0xk4n3ki/restaurant-management-backend/database"
	"github.com/0xk4n3ki/restaurant-management-backend/helpers"
	"github.com/0xk4n3ki/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "null"},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		result, err := userCollection.Aggregate(c, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing users"})
			return
		}

		var allUsers []bson.M
		if err = result.All(c, &allUsers); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while parsing users"})
			return
		}

		if len(allUsers) == 0 {
			ctx.JSON(http.StatusOK, gin.H{"users": []bson.M{}, "total_count": 0})
			return
		}
		ctx.JSON(http.StatusOK, allUsers[0])
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		var user models.User
		userId := ctx.Param("user_id")

		err := userCollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching user"})
			return
		}
		ctx.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		var user models.User

		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(c, bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"phone": user.Phone},
			},
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking user"})
			return
		}

		if count > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email or phone already exists"})
			return
		}

		password, err := HashPassword(*user.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = &password

		user.Created_at = time.Now().UTC()
		user.Updated_at = time.Now().UTC()
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		_, insertErr := userCollection.InsertOne(c, user)
		if insertErr != nil {
			log.Println("error inserting user: ", insertErr)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user not created"})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"user_id": user.User_id})
	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		var user models.User
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var foundUser models.User
		err := userCollection.FindOne(c, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)

		if err := helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tokens"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"user_id":       foundUser.User_id,
			"email":         foundUser.Email,
			"first_name":    foundUser.First_name,
			"last_name":     foundUser.Last_name,
			"token":         token,
			"refresh_token": refreshToken,
		})
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func VerifyPassword(providedPassword string, hashedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
	if err != nil {
		return false, "login or password is incorrect"
	}
	return true, ""
}


func RefreshToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refreshToken := ctx.GetHeader("refresh_token")
		if refreshToken == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "refresh token required"})
			return 
		}

		claims, msg := helpers.ValidateToken(refreshToken)
		if msg != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return 
		}

		newToken, newRefreshToken, _ := helpers.GenerateAllTokens(
			claims.Email,
			claims.First_name,
			claims.Last_name,
			claims.Uid,
		)

		if err := helpers.UpdateAllTokens(newToken, newRefreshToken, claims.Uid); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tokens"})
			return 
		}

		ctx.JSON(http.StatusOK, gin.H{
			"token": newToken,
			"refresh_token": newRefreshToken,
		})
	}
}