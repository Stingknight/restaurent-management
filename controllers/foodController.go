package controllers

import (
	"context"

	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Stingknight/restaurentManagement/database"
	"github.com/Stingknight/restaurentManagement/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var foodCollection *mongo.Collection = database.OpenCollection(database.DBInstance(),"food")



var validate *validator.Validate =validator.New()

func GetFoods() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),10*time.Second)
		defer cancel()

		recordPerPage,err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err!=nil || recordPerPage <1 {
			recordPerPage =10
		}
		page,err := strconv.Atoi(ctx.Query("page"))
		if err!=nil || page < 1{
			page =1
		}
		
		var startIndex int = (page-1)*recordPerPage
		startIndex,_ =strconv.Atoi(ctx.Query("startIndex"))

		matchstage :=bson.D{{Key: "$match",Value: bson.D{{Key:"",Value: ""}}}}
		groupstage :=bson.D{{Key: "$group",Value: bson.D{{Key: "_id",Value: bson.D{{Key:"_id",Value: "null"}}},{Key: "total_count",Value: bson.D{{Key: "$sum",Value: 1}}},{Key: "data",Value: bson.D{{Key: "$push",Value: "$$Root"}}}}}}
		
		projectstage :=bson.D{{Key: "$project",Value: bson.D{{Key:"_id",Value: 1},{Key: "data",Value:bson.D{{Key: "$slice",Value: []interface{}{"$data",startIndex,recordPerPage}}}}}}}

		foodResults,err := foodCollection.Aggregate(context,mongo.Pipeline{matchstage,groupstage,projectstage})
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		var foods []models.Food
		if err = foodResults.All(context,&foods);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
	}
	
}

func GetFood() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		var foodId string =ctx.Param("food_id")
		var food models.Food

		foodObjectId,err := primitive.ObjectIDFromHex(foodId)
		if err!=nil{
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		err = foodCollection.FindOne(context,bson.M{"_id":foodObjectId}).Decode(&food)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,food)
	}
}

func CreateFood() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var menu models.Menu
		var food models.Food

		if err := ctx.BindJSON(&food);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err := validate.Struct(food);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var menu_id string = *food.Menu_id
		menuObjectId,err  :=primitive.ObjectIDFromHex(menu_id)
		if err!=nil{
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		
		if err := menuCollection.FindOne(context,bson.M{"_id":menuObjectId}).Decode(&menu);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		food.Created_at ,_ =time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		food.Updated_at ,_ =time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		food.ID =primitive.NewObjectID()
		food.Food_id =food.ID.Hex()

		var num float64 = toFixed(*food.Price,2)
		food.Price=&num

		insertedResult,err := foodCollection.InsertOne(context,&food)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"food item was not created"})
			return
		}
		
		ctx.IndentedJSON(http.StatusOK,insertedResult)
	}
}

func UpdateFood() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		var menu models.Menu

		var food models.Food
		if err := ctx.BindJSON(&food);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var foodId string =ctx.Param("food_id")

		foodObjectId,err:=primitive.ObjectIDFromHex(foodId)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var filter primitive.M =bson.M{"_id":foodObjectId}
		var updateObj primitive.D

		if food.Name!=nil{
			updateObj = append(updateObj,bson.E{Key:"name",Value: *food.Name})
		}
		if food.Price!=nil{
			updateObj = append(updateObj,bson.E{Key:"price",Value: *food.Price})
		}
		if food.Food_Image!=nil{
			updateObj = append(updateObj,bson.E{Key:"food_image",Value: *food.Food_Image})

		}
		if food.Menu_id!=nil{
			foodMenuObjectId,err := primitive.ObjectIDFromHex(*food.Menu_id)
			if err!=nil{
				ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
				return
			}
			if err = menuCollection.FindOne(context,bson.M{"_id":foodMenuObjectId}).Decode(&menu);err!=nil{
				log.Fatalf("Error: %v",err)
				ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
				return
			}
			updateObj=append(updateObj,bson.E{Key: "menu_id",Value: *food.Menu_id})

			
		}
		food.Updated_at,_=time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		updateObj = append(updateObj,bson.E{Key: "updated_at",Value: food.Updated_at})

		var upsert bool =true

		opts :=options.UpdateOptions{
			Upsert:&upsert ,
		}

		updateResult,err := foodCollection.UpdateOne(context,filter,bson.M{"$set":updateObj},&opts)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,updateResult)
	}
}

func round(num float64) int{
	return int(num+math.Copysign(0.5, num))
}


func toFixed(num float64,precision int) float64{
	output := math.Pow(10,float64(precision))
	return float64(round(num*output))/output
}


// typesense search

// func SearchFood()gin.HandlerFunc{
// 	return func(ctx *gin.Context){
// 		var foodName string= ctx.Param("food_name")
	
// 		searchResults,err := helpers.SearchDocument(foodName)
// 		if err!=nil{
// 			log.Fatalf("Error: %v",err)
// 			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
// 			return
// 		}
		
// 		ctx.IndentedJSON(http.StatusOK,*searchResults)
		
// 	}
// }	