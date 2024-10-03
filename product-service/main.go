package main

import (
	"log"
	"product-service/db"
	"product-service/handler"
	"product-service/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	err := db.Connect("mongodb://localhost:27017", "product-db", "products")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	utils.InitMQ()
	defer utils.CloseMQ()

	router := gin.Default()
	router.POST("/product", handler.CreateProduct)
	router.GET("/product/:name", handler.GetProduct)
	router.GET("/products", handler.GetProducts)
	router.PUT("/product/:name", handler.UpdateInventory)
	router.DELETE("/product/:name", handler.DeleteProduct)
	router.Run(":8082")
}
