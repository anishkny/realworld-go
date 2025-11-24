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
	uuid "github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

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
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=password dbname=postgres port=5432"
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB.AutoMigrate(&User{})
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
