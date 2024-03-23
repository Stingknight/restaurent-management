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
type OrderItemsPack struct {
	Table_id    *string
	Order_items  []models.OrderItem
}  

var orderItemCollection *mongo.Collection =database.OpenCollection(database.DBInstance(),"orderitem")


func GetOrderItems() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		var orderItems []bson.M

		orderItemResult,err := orderItemCollection.Find(context,bson.M{})
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err := orderItemResult.All(context,&orderItems);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while fetching order items"})
			return
		}
		ctx.IndentedJSON(http.StatusBadRequest,orderItems)
	}

}
func GetOrderItemsByOrder() gin.HandlerFunc{
	return func (ctx *gin.Context){
		var orderID string = ctx.Param("order_id")

		allOrderItems, err := ItemsByOrder(orderID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error occured while listing order items by order id"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,allOrderItems)
	}
}

func GetOrderItem() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		var orderItemID string = ctx.Param("order_item__id")

		orderItemObjectID,err := primitive.ObjectIDFromHex(orderItemID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var orderItem models.OrderItem

		if err = orderItemCollection.FindOne(context,bson.M{"_id":orderItemObjectID}).Decode(&orderItem);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,orderItem)
	}
}



func ItemsByOrder(id string)(OrderItems []primitive.M,err error){
	context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

	defer cancel()
	
	matchstage := bson.D{{Key: "$match",Value: bson.D{{Key: "order_id",Value: id}}}}
	// we can do project stage insde pipeline can get required fields of lookedup stage collection or we can do inner lookup stage inside a pipleline of a lookup stage

	// lookupstageFoodex := bson.D{{"$lookup",bson.D{{"from","food"},{"foreignField","food_id"},{"localField","food_id"},{"pipeline",[]interface{}{"$project",bson.D{{"_id",1}}}}}}}

	lookupstageFood := bson.D{{Key: "$lookup",Value: bson.D{{Key: "from",Value: "food"},{Key: "foreignField",Value: "food_id"},{Key: "localField",Value: "food_id"},{Key: "as",Value: "food"}}}}

	unwindstageFood := bson.D{{Key: "$unwind",Value: bson.D{{Key: "path" ,Value: "$food"},{Key: "preserveNullAndEmptyArrays",Value: true}}}}

	// this two below code is exapmle of innerlookup inside a pipleine of lookup
	// lookupstageTableEx := bson.D{{"$lookup",bson.D{{"from","table"},{"foreignField","table_id"},{"localField","table_id"},{"as","table"}}}}

	// lookupstageOrderex := bson.D{{"$lookup",bson.D{{"from","order"},{"foriegnField","order_id"},{"localField","order_id"},{"pipeline",[]interface{}{lookupstageTableEx}}}}}


	lookupstageOrder := bson.D{{Key: "$lookup",Value: bson.D{{Key: "from",Value: "order"},{Key: "foriegnField",Value: "order_id"},{Key: "localField",Value: "order_id"},{Key: "as",Value: "order"}}}}
	unwindstageOrder := bson.D{{Key: "$unwind",Value: bson.D{{Key: "path",Value: "$order"},{Key: "preserveNullAndEmptyArrays",Value: true}}}}


	lookupstageTable := bson.D{{Key: "$lookup",Value: bson.D{{Key: "from",Value: "table"},{Key: "foreignField",Value: "table_id"},{Key: "localField",Value: "order.table_id"},{Key: "as",Value: "table"}}}}
	unwindstagetable := bson.D{{Key: "$unwind",Value: bson.D{{Key: "path",Value: "$table"},{Key: "preserveNullAndEmptyArrays",Value: true}}}}


	projectstage := bson.D{{Key: "$project",Value: bson.D{{Key: "_id",Value: 0},{Key: "amount",Value: "$food.price"},{Key: "total_count",Value: 1},{Key: "food_name",Value: "$food.name"},{Key: "food_image",Value: "$food.food_image"},{Key: "table_number",Value: "$table.table_number"},{Key: "table_id",Value: "$table.table_id"},{Key: "order_id",Value: "$order.order_id"},{Key: "price",Value: "$food.price"},{Key: "quantity",Value: 1}}}}

	gruopstage := bson.D{{Key: "$group",Value: bson.D{{Key: "_id",Value: bson.D{{Key: "order_id",Value: "$order_id"},{Key: "table_id",Value: "$table_id"},{Key: "table_number",Value: "$table_number"}}},{Key: "payment_due",Value: bson.D{{Key: "$sum",Value: "$amount"}}},{Key: "total_count",Value: bson.D{{Key: "$sum",Value: 1}}},{Key: "$order_items",Value: bson.D{{Key: "$push",Value: "$$ROOT"}}}}}}

	projectstage2 :=bson.D{{Key: "$project",Value: bson.D{{Key: "id",Value: 0},{Key: "payment_due",Value: 1},{Key: "total_count",Value: 1},{Key: "table_number",Value: "$_id.table_number"},{Key: "order_items",Value: 1}}}}

	allResults,err:= orderItemCollection.Aggregate(context,mongo.Pipeline{matchstage,lookupstageFood,unwindstageFood,lookupstageOrder,unwindstageOrder,lookupstageTable,unwindstagetable,projectstage,gruopstage,projectstage2})
	if err!=nil{
		log.Fatalf("Error: %v",err)
		return nil, err
		
	}
	if err = allResults.All(context,&OrderItems);err!=nil{
		log.Fatalf("Error: %v",err)
		return nil, err
	}
	return  OrderItems,nil
}

func CreateOrderItem() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var orderItemPack OrderItemsPack
		var order models.Order

		if err := ctx.BindJSON(&orderItemPack);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var orderItemToBeInserted []interface{} =[]interface{}{}

		order.Order_Date,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		order.Table_id = orderItemPack.Table_id
		order_id := OrderItemOrderCreator(order)
		
		for _,orderItem := range orderItemPack.Order_items{

			orderItem.Order_id =&order_id

			if err := validate.Struct(orderItem);err!=nil{
				log.Fatalf("Error: %v",err)
				ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
				return
			}

			orderItem.ID =primitive.NewObjectID()
			orderItem.OrderItem_id =orderItem.ID.Hex()
			orderItem.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
			orderItem.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

			var num float64 =toFixed(*orderItem.Unit_price,2)
			orderItem.Unit_price =&num
			orderItemToBeInserted=append(orderItemToBeInserted,orderItem)
		}
		insertedOrderItem,err := orderItemCollection.InsertMany(context,orderItemToBeInserted)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,insertedOrderItem)
	}
}

func UpdateOrderItem() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel :=context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()
		var orderItem models.OrderItem

		var orderItemID string =ctx.Param("order_item_id")
		if err := ctx.BindJSON(&orderItem);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		orderItemObjectID, err := primitive.ObjectIDFromHex(orderItemID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		var filter primitive.M = bson.M{"_id":orderItemObjectID}

		var updateObj primitive.D

		if orderItem.Unit_price!=nil{
			updateObj = append(updateObj,bson.E{Key: "unit_price",Value: *orderItem.Unit_price})
		}

		if orderItem.Quantity!=nil{
			updateObj = append(updateObj,bson.E{Key: "quantity",Value: *orderItem.Quantity})
		}

		if orderItem.Food_id!=nil{
			updateObj= append(updateObj,bson.E{Key: "food_id",Value: *orderItem.Food_id})
		}

		orderItem.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		updateObj=append(updateObj,bson.E{Key: "updated_at",Value: orderItem.Updated_at})

		var upsert bool =true

		opts := options.UpdateOptions{
			Upsert: &upsert,
		}

		updateResult,err := orderItemCollection.UpdateOne(context,filter,bson.M{"$set":updateObj},&opts)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while updating the orderitem"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,updateResult)
	}
}