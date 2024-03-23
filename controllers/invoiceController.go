package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Stingknight/restaurentManagement/database"
	"github.com/Stingknight/restaurentManagement/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InvoiceViewFormat struct{
	Invoice_id 			string
	Payment_method		string
	Order_id			string
	Payment_status 		*string
	Payment_due			interface{}
	Table_number		interface{}
	Payment_due_date    time.Time
	Order_details		interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.DBInstance(), "invoice")

func GetInvoices() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var invoices []bson.M

		invoiceResults,err := invoiceCollection.Find(context,bson.M{})
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err = invoiceResults.All(context,&invoices);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while geting the invoice results"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,invoices)
	}

}

func GetInvoice() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var invoice models.Invoice

		var paramID string = ctx.Param("invoice_id")

		paramObjectID,err := primitive.ObjectIDFromHex(paramID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err = invoiceCollection.FindOne(context,bson.M{"_id":paramObjectID}).Decode(&invoice);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while geting the invoice result"})
			return
		}
		
		var invoiceView InvoiceViewFormat

		allOrderItems,err := ItemsByOrder(*invoice.Order_id)
		invoiceView.Order_id = *invoice.Order_id
		invoiceView.Payment_due_date =invoice.Payment_due_date

		invoiceView.Payment_method ="null"
		if invoice.Payment_method !=nil{
			invoiceView.Payment_method = *invoice.Payment_method

		}
		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = *&invoice.Payment_status
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number =allOrderItems[0]["tbale_number"]
		invoiceView.Order_details =allOrderItems[0]["order_items"]

		ctx.IndentedJSON(http.StatusOK,invoiceView)
	}
	
}

func CreateInvoice() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var invoice models.Invoice
		if err := ctx.BindJSON(&invoice);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		if err := validate.Struct(invoice);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		var order models.Order

		// var orderID string = *invoice.Order_id
		orderObjectID,err := primitive.ObjectIDFromHex(*invoice.Order_id)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		if err = orderCollection.FindOne(context,bson.M{"_id":orderObjectID}).Decode(&order);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while finding the order"})
			return
		}

		invoice.Payment_due_date,_=time.Parse(time.RFC3339,time.Now().AddDate(0,0,1).Format(time.RFC3339))
		invoice.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		invoice.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		invoice.ID =primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()



		var status string ="PENDING"
		if invoice.Payment_status==nil{
			invoice.Payment_status=&status
		}

		inseterdResult,err :=invoiceCollection.InsertOne(context,&invoice)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,inseterdResult)
	}
}

func UpdateInvoice() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var invoice models.Invoice

		if err := ctx.BindJSON(&invoice);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var invoiceID string = ctx.Param("invoice_id")
		invoiceObjectId,err :=primitive.ObjectIDFromHex(invoiceID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		var filter primitive.M = bson.M{"_id":invoiceObjectId}
		
		var updateObj primitive.D 
		
		if invoice.Payment_method!=nil{
			updateObj=append(updateObj, bson.E{Key: "payment_method",Value: *invoice.Payment_method})
		}

		if invoice.Payment_status!=nil{
			updateObj=append(updateObj, bson.E{Key: "payment_status",Value: *invoice.Payment_status})
		}
		invoice.Updated_at,_=time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		updateObj = append(updateObj,bson.E{Key: "updated_at",Value: invoice.Updated_at})

		var upsert bool =true
		opts :=options.UpdateOptions{
			Upsert: &upsert,
		}

		var status string ="PENDING"
		if invoice.Payment_status==nil{
			invoice.Payment_status=&status
		}
		updateResult,err := invoiceCollection.UpdateOne(context,filter,bson.M{"$set":updateObj},&opts)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while updating hte invoice"})
			return
		}
		
		ctx.IndentedJSON(http.StatusOK,updateResult)
	}
}