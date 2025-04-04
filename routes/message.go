package routes

import (
	"database/sql"
	"log"
	"net/http"

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

		c.IndentedJSON(http.StatusOK, messages)
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

func SetupMessageRoutes(router *gin.Engine, db *sql.DB) {
	messageRoutes := router.Group("/messages")
	{
		messageRoutes.GET("/", GetMessages(db))
		messageRoutes.POST("/", SetMessage(db))
	}
}
