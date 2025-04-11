package main

import (
	"github.com/adrienchanove/alpha-enigma-api/database"
	"github.com/adrienchanove/alpha-enigma-api/routes"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/adrienchanove/alpha-enigma-api/docs"
)

// @title           Enigma chat API
// @version         0.0.1
// @description     API to serve messaging securely
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath  /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer {token}" to correctly authenticate.
// @securitydefinitions.apikey X-User
// @in header
// @name X-User
// @description Type the username to correctly authenticate.

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
