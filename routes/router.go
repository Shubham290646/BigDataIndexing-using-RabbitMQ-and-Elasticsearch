package routes

import (
	"info7255-bigdata-app/database"
	"info7255-bigdata-app/elastic"
	"info7255-bigdata-app/handlers"
	"info7255-bigdata-app/middleware"
	"info7255-bigdata-app/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(gin.Recovery())

	redisRepo := database.NewRedisRepository("localhost:6379")
	planService := services.NewPlanService(redisRepo)
	esFactory := elastic.NewElasticFactory()
	planHandler := handlers.NewPlanHandler(planService, esFactory)

	v1 := router.Group("/v1", middleware.OAuth2Middleware())
	{
		v1.POST("/plan", planHandler.CreatePlan)
		v1.GET("/plan/:objectId", planHandler.GetPlan)
		v1.DELETE("/plan/:objectId", planHandler.DeletePlan)
		v1.PATCH("/plan/:objectId", planHandler.PatchPlan)
		v1.PUT("/plan", planHandler.UpdatePlan)
		v1.GET("/plans", planHandler.GetAllPlans)
		v1.POST("/search", planHandler.SearchPlans)
	}

	return router
}
