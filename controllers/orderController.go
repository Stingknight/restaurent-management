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

var orderCollection *mongo.Collection = database.OpenCollection(database.DBInstance(),"order")

func GetOrders() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var orders []models.Order

		orderResults,err := orderCollection.Find(context,bson.M{})
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while listing all orders"})
			return
		}
		if err = orderResults.All(context,&orders);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,orders)

	}

}

func GetOrder() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var order models.Order

		var orderId string =ctx.Param("order_id")
		orderObjectId,err := primitive.ObjectIDFromHex(orderId)
		if err!=nil{
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		if err = orderCollection.FindOne(context,bson.M{"_id":orderObjectId}).Decode(&order);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while getting the order"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,order)
	}
}

func CreateOrder() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var table models.Table
		var order models.Order

		if err := ctx.BindJSON(&order);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		if err := validate.Struct(order);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var tableID string = *order.Table_id

		tableObjectID,err := primitive.ObjectIDFromHex(tableID)
		if err!=nil{
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err = tableCollection.FindOne(context,bson.M{"_id":tableObjectID}).Decode(&table);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while finding the table"})
			return
		}
		order.Created_at ,_ =time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		order.Updated_at ,_ =time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		order.ID =primitive.NewObjectID()
		order.Order_id =order.ID.Hex()

		insertedResult,err := orderCollection.InsertOne(context,&order)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,insertedResult)

	}
}

func UpdateOrder() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		var table models.Table
		var order models.Order

		if err := ctx.BindJSON(&order);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var orderID string = ctx.Param("order_id")

		orderObjectId,err := primitive.ObjectIDFromHex(orderID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		filter :=bson.M{"_id":orderObjectId}
		var updateObj  primitive.D

		if order.Table_id!=nil{
			orderTableObjectId,err := primitive.ObjectIDFromHex(*order.Table_id)
			if err!=nil{
				ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
				return
			}
			if err = tableCollection.FindOne(context,bson.M{"_id":orderTableObjectId}).Decode(&table);err!=nil{
				log.Fatalf("Error: %v",err)
				ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while finding table data"})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "table_id",Value: *order.Table_id})
		}
		order.Updated_at,_=time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		updateObj = append(updateObj,bson.E{Key: "updated_at",Value: order.Updated_at})

		var upsert bool =true

		opts := options.UpdateOptions{
			Upsert: &upsert,
		}
		updateResult,err := orderCollection.UpdateOne(context,filter,bson.M{"$set":updateObj},&opts)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while updating the order"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,updateResult)
	}
}

func OrderItemOrderCreator(order models.Order) string {
	order.Created_at ,_ =time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	order.Updated_at ,_ =time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

	order.ID = primitive.NewObjectID()
	order.Order_id =order.ID.Hex()

	orderCollection.InsertOne(context.Background(),&order)
	return order.Order_id
}