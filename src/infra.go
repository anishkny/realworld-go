package src

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const JWT_SECRET = "your_secret_key"

var db *gorm.DB

func TrapSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("Shutting down server...")
		os.Exit(0)
	}()
}

type BaseModel struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ConnectDatabase() {
	// Get database DSN from environment variable or use default
	// Format: "host=<host>	user=<user> password=<password> dbname=<dbname> port=<port>"
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=password dbname=postgres port=5432"
	}

	// Create logger to ignore "record not found" errors
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Connect to database
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		//coverage:ignore
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&User{}, &Follow{})
}

func FormatBindErrors(err error) gin.H {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, verr := range validationErrors {
			const sep = "Error:"
			errMsg := verr.Error()
			if idx := strings.Index(errMsg, sep); idx != -1 {
				errMsg = errMsg[idx+len(sep):]
			}
			errors[verr.Field()] = errMsg
		}
	} else {
		errors["error"] = err.Error()
	}
	return gin.H{"errors": errors}
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

func Authenticate(c *gin.Context) {
	// If authorization header is present, validate token and set user in context
	var userId uuid.UUID
	if token := c.GetHeader("Authorization"); token != "" {
		c.Set("token", token)
		userId = ValidateJWT(token)
		if userId == uuid.Nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Set("userId", userId)
	}

	// If no authorization header, check if method+route is allowed unauthenticated
	if userId == uuid.Nil {
		method := c.Request.Method
		route := c.FullPath()
		allowedUnauthenticated := map[string]map[string]bool{
			"GET": {
				"/api":                    true,
				"/api/profiles/:username": true,
			},
			"POST": {
				"/api/users":       true,
				"/api/users/login": true,
			},
		}
		if routes, ok := allowedUnauthenticated[method]; !ok || !routes[route] {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
	}

	c.Next()
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
	tokenString, err := jwt.SignedString([]byte(JWT_SECRET))
	if err != nil {
		//coverage:ignore
		return ""
	}

	return tokenString
}

func ValidateJWT(token string) uuid.UUID {
	// Parse the token
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(JWT_SECRET), nil
	})
	if err != nil || !parsedToken.Valid {
		return uuid.Nil
	}

	// Extract user ID from claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if sub, ok := claims["sub"].(string); ok {
			userId, err := uuid.Parse(sub)
			if err == nil {
				return userId
			}
		}
	}

	//coverage:ignore
	return uuid.Nil
}
