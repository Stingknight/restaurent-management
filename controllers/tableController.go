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

var tableCollection *mongo.Collection = database.OpenCollection(database.DBInstance(), "table")

func GetTables() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		var tables []bson.M

		orderItemResult,err := tableCollection.Find(context,bson.M{})
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err := orderItemResult.All(context,&tables);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while fetching order items"})
			return
		}
		ctx.IndentedJSON(http.StatusBadRequest,tables)
	}

}

func GetTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		var tableID string = ctx.Param("table_id")

		tableObjectID,err := primitive.ObjectIDFromHex(tableID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var table models.Table

		if err = tableCollection.FindOne(context,bson.M{"_id":tableObjectID}).Decode(&table);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		var table models.Table
		if err := ctx.BindJSON(&table);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err := validate.Struct(table);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		table.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		table.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		table.ID =primitive.NewObjectID()
		table.Table_id =table.ID.Hex()
		
		insertedResult,err := tableCollection.InsertOne(context,&table)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,insertedResult)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var table models.Table

		var tableID string = ctx.Param("table_id")

		if err := ctx.BindJSON(&table);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}	
		
		tableObjectID,err := primitive.ObjectIDFromHex(tableID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}	

		var filter primitive.M =bson.M{"_id":tableObjectID}

		var updateObj primitive.D
		
		if table.Number_of_guests!=nil{
			updateObj = append(updateObj,bson.E{Key: "number_of_guests",Value: *table.Number_of_guests})
		}

		if table.Table_number!=nil{
			updateObj = append(updateObj, bson.E{Key: "table_number",Value: *table.Table_number})
		}

		table.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		var upsert bool = true
		opts  :=options.UpdateOptions{
			Upsert: &upsert,
		}

		updateResult, err := tableCollection.UpdateOne(context,filter,bson.M{"$set":updateObj},&opts)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,updateResult)	
	
	}	
}
