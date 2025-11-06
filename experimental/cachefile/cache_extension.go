// karing
package cachefile

import (
	"os"
	"time"

	"github.com/sagernet/bbolt"
	"github.com/sagernet/sing-box/adapter"
)

func (c *CacheFile) BeforeStart() error {
	return c.start(adapter.StartStateInitialize)
}

func (c *CacheFile) DeleteRuleSet(tag string) {
	if c.DB == nil {
		return
	}
	c.DB.View(func(t *bbolt.Tx) error {
		bucket := c.bucket(t, bucketRuleSet)
		if bucket == nil {
			return os.ErrNotExist
		}
		bucket.Delete([]byte(tag))
		return nil
	})
}

func (c *CacheFile) HasRuleSet(tag string) bool {
	if c.DB == nil {
		return false
	}
	err := c.DB.View(func(t *bbolt.Tx) error {
		bucket := c.bucket(t, bucketRuleSet)
		if bucket == nil {
			return os.ErrNotExist
		}
		setBinary := bucket.Get([]byte(tag))
		if len(setBinary) == 0 {
			return os.ErrInvalid
		}
		return nil
	})
	return err == nil
}

func (c *CacheFile) GetAllRuleSetLastUpdated() map[string]time.Time {
	keys := make(map[string]time.Time)
	if c.DB == nil {
		return keys
	}
	c.DB.View(func(t *bbolt.Tx) error {
		bucket := c.bucket(t, bucketRuleSet)
		if bucket == nil {
			return os.ErrNotExist
		}
		bucket.ForEach(func(name []byte, setBinary []byte) error {
			if len(name) != 0 {
				var savedSet adapter.SavedBinary
				savedSet.UnmarshalBinary(setBinary)
				keys[string(name)] = savedSet.LastUpdated
			}
			return nil
		})
		return nil
	})

	return keys
}

func (c *CacheFile) GetAllRuleSetFetchError() map[string]string {
	return c.fetchError
}

func (c *CacheFile) SetRulesetFetchError(tag string, err string) {
	if len(err) == 0 {
		delete(c.fetchError, tag)
	} else {
		c.fetchError[tag] = err
	}
}
