package src

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	err = gorm.G[User](DB).Create(context.Background(), &user)
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

func GenerateJWT(user User) string {
	// Create the JWT claims, which includes the username and expiry time
	claims := jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	// Create token
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret
	secret := []byte("your_secret_key") // Replace with your actual secret key
	tokenString, err := jwt.SignedString(secret)
	if err != nil {
		//coverage:ignore
		return ""
	}

	return tokenString
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
