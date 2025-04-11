package routes

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"slices"

	"github.com/gin-gonic/gin"
)

type Message struct {
	ID         int    `json:"id"`
	Content    string `json:"content"`
	SenderId   int    `json:"senderId"`
	ReceiverId int    `json:"receiverId"`
}

// getMessages godoc
// @Summary Get all messages
// @Description Get a list of all messages
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {array} Message
// @Security ApiKeyAuth
// @Security X-User
// @Router /messages [get]
func GetMessages(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, content, sender_id, receiver_id FROM messages")
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
			return
		}
		defer rows.Close()

		var messages []Message
		for rows.Next() {
			var m Message
			if err := rows.Scan(&m.ID, &m.Content, &m.SenderId, &m.ReceiverId); err != nil {
				log.Println(err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
				return
			}
			messages = append(messages, m)
		}

		if len(messages) == 0 {
			c.IndentedJSON(http.StatusOK, []Message{})
		} else {
			c.IndentedJSON(http.StatusOK, messages)
		}

	}
}

// setMessage godoc
// @Summary Create a new message
// @Description Create a new message with the input payload
// @Tags messages
// @Accept json
// @Produce json
// @Param message body Message true "Create message"
// @Success 200 {object} Message
// @Security ApiKeyAuth
// @Security X-User
// @Router /messages [post]
func SetMessage(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newMessage Message
		if err := c.ShouldBindJSON(&newMessage); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if newMessage.Content == "" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "content is required"})
			return
		}
		// Check if senderId and receiverId are valid
		if newMessage.SenderId <= 0 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "senderId must be a positive integer"})
			return
		}
		if newMessage.ReceiverId <= 0 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "receiverId must be a positive integer"})
			return
		}

		result, err := db.Exec("INSERT INTO messages (content, sender_id, receiver_id) VALUES (?, ?, ?)", newMessage.Content, newMessage.SenderId, newMessage.ReceiverId)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}
		newMessage.ID = int(id)

		c.IndentedJSON(http.StatusCreated, newMessage)
	}
}

// getDiscussions godoc
// @Summary Return a list of user who a user speaks to
// @Description Return a user list who the user have discussed with
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {array} UserGet
// @Security ApiKeyAuth
// @Security X-User
// @Router /messages/getDiscussions/ [get]
func GetDiscussions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("X-User")
		if username == "" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "X-User header is required"})
			return
		}
		var userId int
		var err error
		userId, err = getUserIDByUsername(db, username)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}

		if userId <= 0 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "userId must be a positive integer"})
			return
		}
		rows, err := db.Query("SELECT DISTINCT receiver_id, sender_id FROM messages WHERE sender_id = ? or receiver_id = ?", userId, userId)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get discussions"})
			return
		}
		defer rows.Close()

		// userIds est une liste de int
		var userIds []int

		// length rows
		for rows.Next() {
			var u int
			var u2 int
			if err := rows.Scan(&u, &u2); err != nil {
				log.Println(err)
			}

			if !slices.Contains(userIds, u) && u != userId {
				userIds = append(userIds, u)
			}

			if !slices.Contains(userIds, u2) && u2 != userId {
				userIds = append(userIds, u2)
			}

		}

		var users []UserGet

		for _, userId := range userIds {
			var u UserGet

			err := db.QueryRow("SELECT id, username FROM users WHERE id = ?", userId).Scan(&u.ID, &u.Username)
			if err != nil {
				if err == sql.ErrNoRows {
					continue
				} else {
					log.Println(err)
					c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
					return
				}
			}
			users = append(users, u)
		}

		if len(users) == 0 {
			c.IndentedJSON(http.StatusOK, []UserGet{})
		} else {
			c.IndentedJSON(http.StatusOK, users)
		}
	}
}

// getMessageWith godoc
// @Summary get messages with a user
// @Description get messages with a user
// @Tags messages
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {array} Message
// @Security ApiKeyAuth
// @Security X-User
// @Router /messages/getMessagesWith/{userId} [get]
func GetMessagesWith(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		foreignUserIdStr := c.Param("userId")
		foreignUserId, err := strconv.Atoi(foreignUserIdStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		if foreignUserId <= 0 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "userId must be a positive integer"})
			return
		}
		username := c.GetHeader("X-User")
		if username == "" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "X-User header is required"})
			return
		}
		userId, err := getUserIDByUsername(db, username)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}

		rows, err := db.Query("SELECT id, content, sender_id, receiver_id FROM messages WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userId, foreignUserId, foreignUserId, userId)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
			return
		}
		defer rows.Close()

		var messages []Message
		for rows.Next() {
			var m Message
			if err := rows.Scan(&m.ID, &m.Content, &m.SenderId, &m.ReceiverId); err != nil {
				log.Println(err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
				return
			}
			messages = append(messages, m)
		}

		if len(messages) == 0 {
			c.IndentedJSON(http.StatusOK, []Message{})
		} else {
			c.IndentedJSON(http.StatusOK, messages)
		}
	}
}

func SetupMessageRoutes(router *gin.Engine, db *sql.DB) {
	messageRoutes := router.Group("/messages")
	{
		messageRoutes.GET("/", GetMessages(db))
		messageRoutes.POST("/", SetMessage(db))
		messageRoutes.GET("/getDiscussions/", GetDiscussions(db))
		messageRoutes.GET("/getMessagesWith/:userId", GetMessagesWith(db))

	}
}

// getUserIDByUsername
func getUserIDByUsername(db *sql.DB, username string) (int, error) {
	var userId int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("user not found")
		} else {
			return 0, err
		}
	}
	return userId, nil
}
