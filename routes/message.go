package routes

import (
	"database/sql"
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
// @Param userId path int true "User ID"
// @Success 200 {array} User
// @Router /messages/getDiscussions/{userId} [get]
func GetDiscussions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdStr := c.Param("userId")
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
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
		log.Println(rows.Columns())

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

		var users []User

		for _, userId := range userIds {
			var u User

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
			c.IndentedJSON(http.StatusOK, []User{})
		} else {
			c.IndentedJSON(http.StatusOK, users)
		}
	}
}

// getMessageBetween godoc
// @Summary get messages between two user
// @Description get messages between two user
// @Tags messages
// @Accept json
// @Produce json
// @Param userId1 path int true "User ID 1"
// @Param userId2 path int true "User ID 2"
// @Success 200 {array} Message
// @Router /messages/getMessagesBetween/{userId1}/{userId2} [get]
func GetMessagesBetween(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId1Str := c.Param("userId1")
		userId2Str := c.Param("userId2")

		userId1, err := strconv.Atoi(userId1Str)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID 1"})
			return
		}
		userId2, err := strconv.Atoi(userId2Str)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID 2"})
			return
		}
		if userId1 <= 0 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "userId1 must be a positive integer"})
			return
		}
		if userId2 <= 0 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "userId2 must be a positive integer"})
			return
		}

		rows, err := db.Query("SELECT id, content, sender_id, receiver_id FROM messages WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userId1, userId2, userId2, userId1)
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
		messageRoutes.GET("/getDiscussions/:userId", GetDiscussions(db))
		messageRoutes.GET("/getMessagesBetween/:userId1/:userId2", GetMessagesBetween(db))
	}
}
