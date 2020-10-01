package main

import (
	"fmt"
	"tripify-backend/gin"
)

func main() {
	fmt.Println("Running server on ")
	r := gin.SetupRouter()
	r.Run(":12000")
}