package src

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
	"gorm.io/gorm"
)

func GetProfile(c *gin.Context) {
	followedUsername := c.Param("username")

	// Retrieve followedUser by username
	followedUser, err := gorm.G[User](db).Where("username = ?", followedUsername).First(c)
	if err == gorm.ErrRecordNotFound {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Return profile response
	c.JSON(200, CreateProfileResponse(c, followedUser))
}

func FollowUser(c *gin.Context) {
	var followerUserId uuid.UUID
	if val, exists := c.Get("userId"); exists {
		followerUserId = val.(uuid.UUID)
	}
	followedUsername := c.Param("username")

	// Retrieve followedUser by username
	followedUser, err := gorm.G[User](db).Where("username = ?", followedUsername).First(c)
	if err == gorm.ErrRecordNotFound {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Create follow relationship
	follow := Follow{
		FollowerID: followerUserId,
		FollowedID: followedUser.Id,
	}
	err = gorm.G[Follow](db).Create(c, &follow)
	if err != nil {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to follow user"})
		return
	}

	// Return profile response
	c.JSON(200, CreateProfileResponse(c, followedUser))
}

func UnfollowUser(c *gin.Context) {
	var followerUserId uuid.UUID
	if val, exists := c.Get("userId"); exists {
		followerUserId = val.(uuid.UUID)
	}
	followedUsername := c.Param("username")

	// Retrieve followedUser by username
	followedUser, err := gorm.G[User](db).Where("username = ?", followedUsername).First(c)
	if err == gorm.ErrRecordNotFound {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Delete follow relationship
	_, err = gorm.G[Follow](db).Where("follower_id = ? AND followed_id = ?", followerUserId, followedUser.Id).Delete(c)
	if err != nil {
		//coverage:ignore
		c.JSON(500, gin.H{"error": "Failed to unfollow user"})
		return
	}

	// Return profile response
	c.JSON(200, CreateProfileResponse(c, followedUser))
}

func CreateProfileResponse(c *gin.Context, followedUser User) ProfileResponseEnvelope {
	var isFollowing bool = false

	// Get followerUserId from context if available
	var followerUserId uuid.UUID
	if val, exists := c.Get("userId"); exists {
		// Check if follower is following followedUser
		followerUserId = val.(uuid.UUID)
		follow, err := gorm.G[Follow](db).Where("follower_id = ? AND followed_id = ?", followerUserId, followedUser.Id).First(c)
		if err == nil && follow.Id != uuid.Nil {
			isFollowing = true
		} else if err != gorm.ErrRecordNotFound {
			//coverage:ignore
			c.JSON(500, gin.H{"error": "Failed to check following status"})
			return ProfileResponseEnvelope{}
		}
	}

	return ProfileResponseEnvelope{Profile: ProfileResponse{
		Username:  followedUser.Username,
		Bio:       followedUser.Bio,
		Image:     followedUser.Image,
		Following: isFollowing,
	}}
}

type ProfileResponseEnvelope struct {
	Profile ProfileResponse `json:"profile"`
}

type ProfileResponse struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

type Follow struct {
	BaseModel
	FollowerID uuid.UUID `gorm:"type:uuid;not null;index"`
	FollowedID uuid.UUID `gorm:"type:uuid;not null;index"`
}
