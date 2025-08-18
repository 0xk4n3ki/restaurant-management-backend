package helpers

import (
	"context"
	"os"
	"time"

	"github.com/0xk4n3ki/restaurant-management-backend/database"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type SignedDetails struct {
	Email string
	First_name string
	Last_name string
	Uid string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email, firstName, lastName, uid string) (signedToken, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email: email,
		First_name: firstName,
		Last_name: lastName,
		Uid: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(30*time.Minute).Unix(),
		},
	}
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24*time.Hour).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func UpdateAllTokens(signedToken, signedRefreshToken, userId string) error {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateObj := bson.M{
		"token": signedToken,
		"refresh_token": signedRefreshToken,
		"updated_at": time.Now().UTC(),
	}

	filter := bson.M{"user_id": userId}
	upsert := true
	opts := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		c,
		filter,
		bson.M{"$set": updateObj},
		&opts,
	)

	return err
}

func ValidateToken(signedToken string) (*SignedDetails, string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		return nil, err.Error()
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok || !token.Valid {
		return nil, "the token is invalid"
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, "the token is expired"
	}

	return claims, ""
}