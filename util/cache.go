package util

import (
    "sync"
    "time"
    "videohub/model"
    "videohub/config"
)

type Cache struct {
    videoMap map[string]*cacheItem
    mu       sync.RWMutex
}

type cacheItem struct {
    video *model.Video
    timestamp time.Time
}

func NewCache() *Cache {
    return &Cache{
        videoMap: make(map[string]*cacheItem),
    }
}

func (c *Cache) Get(key string) (*model.Video, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    c.cleanup()

    item, exists := c.videoMap[key]
    if !exists {
        return nil, false
    }

    if time.Since(item.timestamp) > config.Config.CacheRetentionTime * time.Hour {
        return nil, false
    }

    return item.video, true
}

func (c *Cache) Set(key string, video *model.Video) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.videoMap[key] = &cacheItem{
        video: video,
        timestamp: time.Now(),
    }
}

func (c *Cache) cleanup() {
    for key, item := range c.videoMap {
        if time.Since(item.timestamp) > config.Config.CacheRetentionTime * time.Hour {
            delete(c.videoMap, key)
        }
    }
}