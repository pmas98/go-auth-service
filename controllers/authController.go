package controllers

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/pmas98/go-auth-service/config"
	"github.com/pmas98/go-auth-service/models"
	"github.com/pmas98/go-auth-service/utils"
	"golang.org/x/crypto/bcrypt"
)

var db *gorm.DB
var JWTKey []byte

func init() {
	config.Connect()
	db = config.GetDB()
	models.AutoMigrate(db)

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	encodedKey := os.Getenv("EncodedKey")
	JWTKey, _ = base64.StdEncoding.DecodeString(encodedKey)

	// Initialize Kafka producer
	err := utils.InitKafkaProducer()
	if err != nil {
		panic(err) // Handle error appropriately in your application startup
	}
}

// HealthCheck godoc
// @Summary      Health Check
// @Description  Get the health status of the service and its dependencies (database and Redis).
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      503  {object}  map[string]interface{}
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	// Check database connection
	dbErr := db.DB().Ping()

	kafkaErr := utils.SendMessageToKafka("test-topic", "Ping request!", "ping")

	// Prepare response
	response := gin.H{
		"status":    "up",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": gin.H{
			"database": "up",
			"kafka":    "up",
		},
	}

	statusCode := http.StatusOK

	// Check for errors and update response accordingly
	if dbErr != nil {
		response["services"].(gin.H)["database"] = "down"
		response["status"] = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	if kafkaErr != nil {
		response["services"].(gin.H)["kafka"] = "down"
		response["status"] = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	// Send response
	c.JSON(statusCode, response)
}

func HandleSignup(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if user with the same email already exists
	var existingUser models.User
	if err := config.DB.Where("email = ?", newUser.Email).First(&existingUser).Error; err == nil {
		// User with this email already exists
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists with this email"})
		return
	}

	// Hash password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	newUser.Password = string(hashedPassword)

	// Create user in PostgreSQL
	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

func HandleLogin(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Query user from PostgreSQL
	var storedUser models.User
	if err := config.DB.Where("email = ?", user.Email).First(&storedUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Login attempt failed: User not found for email: %s", user.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		} else {
			log.Printf("Database error during login: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred during login"})
		}
		return
	}

	// Compare stored hashed password with input password
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		log.Printf("Login attempt failed: Incorrect password for email: %s", user.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = storedUser.ID
	claims["name"] = storedUser.Name
	claims["email"] = storedUser.Email
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expires in 24 hours
	tokenString, err := token.SignedString(JWTKey)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func DeleteAll(c *gin.Context) {
	// Perform the deletion
	result := config.DB.Unscoped().Delete(&models.User{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ("Deleted all users")})
}
