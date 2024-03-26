package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Stingknight/restaurentManagement/database"
	"github.com/Stingknight/restaurentManagement/helpers"
	"github.com/Stingknight/restaurentManagement/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.DBInstance(),"user")

func GetUsers() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel := context.WithTimeout(context.Background(),10*time.Second)

		defer  cancel()
		var allUsers []bson.M
		
		recordPerPage,err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err!=nil || recordPerPage <1 {
			recordPerPage =10
		}
		page,err :=strconv.Atoi(ctx.Query("page"))
		if err!=nil || page < 1{
			page =1
		}
		var startIndex int = (page-1)*recordPerPage
		startIndex,_ =strconv.Atoi(ctx.Query("startIndex"))
		
		matchstage := bson.D{{Key: "$match",Value: bson.D{{Key: "",Value: ""}}}}
		projectstage := bson.D{{"$project",bson.D{{"_id",0},{"total_count",1},{"user_items",bson.D{{"$slice",[]interface{}{"$data",startIndex,recordPerPage}}}}}}}
		
		allUsersResult,err := userCollection.Aggregate(context,mongo.Pipeline{matchstage,projectstage})
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err = allUsersResult.All(context,&allUsers);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while fetching all users"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,allUsers)
	}
}

func GetUser() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel := context.WithTimeout(context.Background(),10*time.Second)

		defer  cancel()

		var user models.User

		var userID string = ctx.Param("user_id")
		userObjectID,err := primitive.ObjectIDFromHex(userID)
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		if err = userCollection.FindOne(context,bson.M{"_id":userObjectID}).Decode(&user);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while fetching user"})
			return
		}
		ctx.IndentedJSON(http.StatusOK,user)
	}
}

func SignUp() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel := context.WithTimeout(context.Background(),10*time.Second)

		defer  cancel()

		var user models.User
		if err := ctx.BindJSON(&user);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		if err := validate.Struct(user);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		
		count,err := userCollection.CountDocuments(context,bson.M{"$or":[]bson.M{{"email":*user.Email},{"phone":*user.Phone}}})
		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}

		if count> 0{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"user already exists"})
			return
		}
		
		hashed_password := Hashpassword(*user.Password)
		user.Password=&hashed_password
		user.ID =primitive.NewObjectID()
		user.User_id =user.ID.Hex()
		user.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		user.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		insertedResult,err := userCollection.InsertOne(context,&user)

		if err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		ctx.IndentedJSON(http.StatusOK,insertedResult)
	}
}

func Login() gin.HandlerFunc{
	return func (ctx *gin.Context){
		context,cancel := context.WithTimeout(context.Background(),10*time.Second)

		defer  cancel()
	
		var user models.User
		if err := ctx.BindJSON(&user);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":err})
			return
		}
		var foundUser models.User
		if err := userCollection.FindOne(context,bson.M{"email":*user.Email}).Decode(&foundUser);err!=nil{
			log.Fatalf("Error: %v",err)
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":"error while finding user"})
			return
		}
		
		passwordValid,msg := VerifyPassword(*foundUser.Password,*user.Password)
		if !passwordValid{
			ctx.IndentedJSON(http.StatusBadRequest,gin.H{"error":msg})
			return
		}

		tokenString := helpers.GenerateAllToken(foundUser.ID)
		// ctx.SetSameSite(http.SameSiteLaxMode)
		// ctx.SetCookie("Authorization", tokenString, 3600, "", "", false, true)
		ctx.IndentedJSON(http.StatusOK, gin.H{"access_token":tokenString})


	}
}
// hash-password
func Hashpassword(password string)string{
	hashed_password,_ := bcrypt.GenerateFromPassword([]byte(password),0)
	return string(hashed_password)
}
// verify hasing password
func VerifyPassword(hashPassword string,password string)(bool,string){
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword),[]byte(password))

	var check bool =true
	if err!=nil{
		check=false
		return check,fmt.Sprint("password is incorrect")
	}
	return check,fmt.Sprint("password is correct")
}

