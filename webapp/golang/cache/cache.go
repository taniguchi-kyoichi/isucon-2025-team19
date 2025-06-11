package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/catatsuy/private-isu/webapp/golang/models"
)

var (
	client *memcache.Client
)

const (
	userCacheKeyPrefix    = "user_"
	categoryCacheKeyPrefix = "category_"
	defaultExpiration     = 3600 // 1 hour
)

func Initialize(memcacheAddress string) {
	client = memcache.New(memcacheAddress)
	// Start the metrics reporter
	StartCacheMetricsReporter()
}

func GetClient() *memcache.Client {
	return client
}

// User cache functions
func GetUserFromCache(userID int) (*models.User, error) {
	key := fmt.Sprintf("%s%d", userCacheKeyPrefix, userID)
	item, err := client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			recordUserCacheMiss()
			return nil, nil
		}
		return nil, err
	}

	var user models.User
	err = json.Unmarshal(item.Value, &user)
	if err != nil {
		return nil, err
	}

	recordUserCacheHit()
	return &user, nil
}

func SetUserCache(user *models.User) error {
	if user == nil {
		return nil
	}

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%d", userCacheKeyPrefix, user.ID)
	return client.Set(&memcache.Item{
		Key:        key,
		Value:      data,
		Expiration: int32(defaultExpiration),
	})
}

func DeleteUserCache(userID int) error {
	key := fmt.Sprintf("%s%d", userCacheKeyPrefix, userID)
	return client.Delete(key)
}

// Category cache functions
func GetCategoryFromCache() ([]models.Category, error) {
	item, err := client.Get(categoryCacheKeyPrefix)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			recordCatCacheMiss()
			return nil, nil
		}
		return nil, err
	}

	var categories []models.Category
	err = json.Unmarshal(item.Value, &categories)
	if err != nil {
		return nil, err
	}

	recordCatCacheHit()
	return categories, nil
}

func SetCategoryCache(categories []models.Category) error {
	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}

	return client.Set(&memcache.Item{
		Key:        categoryCacheKeyPrefix,
		Value:      data,
		Expiration: int32(defaultExpiration),
	})
}

func DeleteCategoryCache() error {
	return client.Delete(categoryCacheKeyPrefix)
} 