package controllers

import (
	"context"
	"fmt"
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
var menuCollection *mongo.Collection =database.OpenCollection(database.DBInstance(),"menu")

func GetMenus() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var menus []bson.M

		menuResults,err := menuCollection.Find(context,bson.M{})
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err = menuResults.All(context,&menus);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error occured while fetching menu"})
			return
		}
		fmt.Println(menus)
		ctx.IndentedJSON(http.StatusOK,menus)
	}

}

func GetMenu() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		var menuId string =ctx.Param("menu_id")
		var menu models.Menu

		menuObjectId,err := primitive.ObjectIDFromHex(menuId)
		if err!=nil{
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		if err = menuCollection.FindOne(context,bson.M{"_id":menuObjectId}).Decode(&menu);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error occured while fetching menu"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,menu)
	}
}

func CreateMenu() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var menu models.Menu 

		if err := ctx.BindJSON(&menu);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		if err := validate.Struct(menu);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		menu.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		menu.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

		menu.ID =primitive.NewObjectID()
		menu.Menu_id =menu.ID.Hex()

		insertedResult,err:= menuCollection.InsertOne(context,&menu)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error occured while inserting menu"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,insertedResult)
	}
}

func UpdateMenu() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var menu models.Menu

		if err := ctx.BindJSON(&menu);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var menuId string =ctx.Param("menu_id")

		menuObjectId,err :=primitive.ObjectIDFromHex(menuId)
		if err!=nil{
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		} 
		filter :=bson.M{"_id":menuObjectId}
		
		var updateObj primitive.D

		if menu.Start_Date !=nil && menu.End_Date !=nil{
			if !inTimespan(*menu.Start_Date,*menu.End_Date,time.Now()){
				log.Fatalf("Error: %v",err)
				ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
				return
			}

			updateObj =append(updateObj,bson.E{Key: "start_date",Value: *menu.Start_Date})
			updateObj = append(updateObj,bson.E{Key: "end_date",Value: *menu.End_Date})

			if *menu.Name != ""{
				updateObj = append(updateObj,bson.E{Key: "name",Value: *menu.Name})
			}
			if *menu.Category !=""{
				updateObj= append(updateObj,bson.E{Key: "category",Value: *menu.Category})
			}
			menu.Updated_at,_=time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
			updateObj = append(updateObj,bson.E{Key: "updated_at",Value: menu.Updated_at})

			var upsert bool =true
			opt :=options.UpdateOptions{
				Upsert:&upsert,
			}

			updateResult,err := menuCollection.UpdateOne(context,filter,bson.M{"$set":updateObj},&opt)
			if err!=nil{
				log.Fatalf("Error: %v",err)
				ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error ocurred while updating the menu"})
				return
			}
			ctx.IndentedJSON(http.StatusOK,updateResult)
		}
		
	}
}

func inTimespan(start,end,check time.Time)bool{
	return start.After(time.Now()) && end.After(start)
}