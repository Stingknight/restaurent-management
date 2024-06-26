package main

import (

	"os"
	"github.com/Stingknight/restaurentManagement/middleware"
	"github.com/Stingknight/restaurentManagement/routes"
	"github.com/gin-gonic/gin"
	
)


func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()

	router.Use(gin.Logger())
	routes.UserRoutes(router)
	// router for authentication
	router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)
	// running the port in 9000
	router.Run("localhost:9000")

}
