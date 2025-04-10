package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserGet struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	PublicKey string `json:"publicKey"`
}

type UserPost struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	PublicKey string `json:"publicKey"`
}

// getUsers godoc
// @Summary Get all users
// @Description Get a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} UserGet
// @Router /users [get]
func GetUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, username FROM users")
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
			return
		}
		defer rows.Close()

		var users []UserGet
		for rows.Next() {
			var u UserGet
			if err := rows.Scan(&u.ID, &u.Username); err != nil {
				log.Println(err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
				return
			}
			users = append(users, u)
		}

		c.IndentedJSON(http.StatusOK, users)
	}
}

// setUser godoc
// @Summary Create a new user
// @Description Create a new user with the input payload
// @Tags users
// @Accept json
// @Produce json
// @Param user body UserPost true "Create user"
// @Success 201 {object} UserGet
// @Router /users [post]
func SetUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newUser UserPost
		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if newUser.Username == "" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}
		if newUser.PublicKey == "" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "publicKey is required"})
			return
		}

		result, err := db.Exec("INSERT INTO users (username, public_key) VALUES (?, ?)", newUser.Username, newUser.PublicKey)

		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		newUser.ID = int(id)

		c.IndentedJSON(http.StatusCreated, newUser)
	}
}

// GetUserById godoc
// @Summary Get a user by ID
// @Description Get a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserGet
// @Router /users/{id} [get]
func GetUserById(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		inputId := c.Param("id")

		id, err := strconv.Atoi(inputId)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var user UserGet
		err = db.QueryRow("SELECT id, username FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			} else {
				log.Println(err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			}
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func SetupUserRoutes(router *gin.Engine, db *sql.DB) {
	userRoutes := router.Group("/users")
	{
		userRoutes.GET("/", GetUsers(db))
		userRoutes.GET("/:id", GetUserById(db))
		
	}
}


func SetupPublicUserRoutes(router *gin.Engine, db *sql.DB) {
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("/", SetUser(db))
	}
}