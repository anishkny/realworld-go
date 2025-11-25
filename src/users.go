package src

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RegisterUser(c *gin.Context) {
	var userDTOEnv UserRegistrationDTOEnvelope
	var err error
	if err = c.ShouldBindJSON(&userDTOEnv); err != nil {
		c.JSON(422, FormatBindErrors(err))
		return
	}
	userDTO := userDTOEnv.User

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := User{
		Email:    userDTO.Email,
		Username: userDTO.Username,
		Password: string(hashedPassword),
	}
	err = gorm.G[User](db).Create(c, &user)
	if err != nil {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token := GenerateJWT(user)
	if token == "" {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(200, gin.H{"user": CreateUserResponse(user, token)})
}

func LoginUser(c *gin.Context) {
	var userDTOEnv UserLoginDTOEnvelope
	var err error
	if err = c.ShouldBindJSON(&userDTOEnv); err != nil {
		c.JSON(422, FormatBindErrors(err))
		return
	}
	userDTO := userDTOEnv.User

	// Find user by email
	user, err := gorm.G[User](db).Where("email = ?", userDTO.Email).First(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "User not found"})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userDTO.Password))
	if err != nil {
		c.JSON(401, gin.H{"error": "Wrong password"})
		return
	}

	// Generate JWT token
	token := GenerateJWT(user)
	if token == "" {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(200, gin.H{"user": CreateUserResponse(user, token)})
}

func GetUser(c *gin.Context) {
	userId, _ := c.Get("userId")
	token, _ := c.Get("token")

	// Find user by ID
	user, err := gorm.G[User](db).Where("id = ?", userId.(uuid.UUID)).First(c)
	if err != nil {
		//coverage:ignore
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Return response
	c.JSON(200, gin.H{"user": CreateUserResponse(user, token.(string))})
}

// ---------- Helpers ---------- //
func CreateUserResponse(user User, token string) UserResponse {
	return UserResponse{
		Email:    user.Email,
		Username: user.Username,
		Token:    token,
		Bio:      user.Bio,
		Image:    user.Image,
	}
}

// ---------- DTOs ---------- //
type UserRegistrationDTOEnvelope struct {
	User UserRegistrationDTO `json:"user"`
}

type UserRegistrationDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginDTOEnvelope struct {
	User UserLoginDTO `json:"user"`
}

type UserLoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponseEnvelope struct {
	User UserResponse `json:"user"`
}

type UserResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

// ---------- Model ---------- //
type User struct {
	BaseModel
	Email    string `gorm:"uniqueIndex;not null"`
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Bio      string `gorm:"type:text;not null;default:''"`
	Image    string `gorm:"type:text;not null;default:''"`
}
