package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/0xk4n3ki/restaurant-management-backend/database"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type InvoiceViewFormat struct {
	Invoice_id       string
	Payment_method   string
	Order_id         string
	Payment_status   *string
	Payment_due      interface{}
	Table_number     interface{}
	Payment_due_date time.Time
	Order_details    interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := invoiceCollection.Find(c, bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing invoice items"})
			return 
		}

		var allInvoices []bson.M
		if err = result.All(c, &allInvoices); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
