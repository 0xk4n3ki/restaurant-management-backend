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
		c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
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
		c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		invoiceId := ctx.Param("invoice_id")
		var invoice models.Invoice
		err := invoiceCollection.FindOne(c, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing invoice item"})
			return
		}

		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.Order_id)

		invoiceView.Order_id = invoice.Order_id
		invoiceView.Payment_due_date = invoice.Payment_due_date
		invoiceView.Payment_method = "null"
		if invoice.Payment_method != nil {
			invoiceView.Payment_method = *invoice.Payment_method
		}
		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = invoice.Payment_status
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_items"]

		ctx.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		var invoice models.Invoice
		if err := ctx.BindJSON(&invoice); err != nil {
			return
		}

		var order models.Order
		err := orderCollection.FindOne(c, bson.M{"order_id": invoice.Order_id}).Decode(&order)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "order was not found"})
			return
		}

		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}

		invoice.Payment_due_date = time.Now().UTC().AddDate(0, 0, 1)
		invoice.Created_at = time.Now().UTC()
		invoice.Updated_at = time.Now().UTC()

		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := invoiceCollection.InsertOne(c, invoice)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invoice item not created"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
		defer cancel()

		var invoice models.Invoice
		invoiceId := ctx.Param("invoice_id")

		if err := ctx.BindJSON(&invoice); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D
		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_method", Value: invoice.Payment_method})
		}
		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.Payment_status})
		}

		invoice.Updated_at = time.Now().UTC()
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: invoice.Updated_at})

		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}
		updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.Payment_status})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := invoiceCollection.UpdateOne(
			c,
			bson.M{"invoice_id": invoiceId},
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invoice item update failed"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
