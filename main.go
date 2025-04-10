package main

import (
	"github.com/adrienchanove/alpha-enigma-api/database"
	"github.com/adrienchanove/alpha-enigma-api/routes"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/adrienchanove/alpha-enigma-api/docs"
)

// @title           My Test Api
// @version         1.0
// @description     My Test Api to compare go and node
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

//	@_host      localhost:8080
//
// @BasePath  /
// @_schemes http
func main() {
	router := gin.Default()

	// Ajout de la route pour Swagger UI
	router.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// / redirect to /doc/index.html
	router.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/doc/index.html")
	})
	// 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	// Initialize the database
	database.InitDB("./alpha-enigma.db")
	db := database.DB

	// public routes
	routes.SetupAuthRoutes(router, db)
	routes.SetupPublicUserRoutes(router, db)

	// private routes
	router.Use(routes.AuthMiddleware())

	routes.SetupUserRoutes(router, db)
	routes.SetupMessageRoutes(router, db)

	router.Run("localhost:8080")
}
