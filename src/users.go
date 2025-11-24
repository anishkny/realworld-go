package src

import (
	"context"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RegisterUser(c *gin.Context) {
	var userDTOEnv UserRegistrationDTOEnvelope
	if err := c.ShouldBindJSON(&userDTOEnv); err != nil {
		c.JSON(422, FormatBindErrors(err))
		return
	}
	userDTO := userDTOEnv.User

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	user := User{
		Email:    userDTO.Email,
		Username: userDTO.Username,
		Password: string(hashedPassword),
	}

	err2 := gorm.G[User](DB).Create(context.Background(), &user)
	if err2 != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(200, gin.H{"user": UserResponse{
		Email:    user.Email,
		Username: user.Username,
		Token:    "dummy-token",
		Bio:      user.Bio,
		Image:    user.Image,
	}})
}

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

type User struct {
	BaseModel
	Email    string `gorm:"uniqueIndex;not null"`
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Bio      string `gorm:"type:text;not null;default:''"`
	Image    string `gorm:"type:text;not null;default:''"`
}
