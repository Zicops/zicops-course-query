package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetTopicResources(ctx context.Context, topicID *string) ([]*model.TopicResource, error) {
	currentResources := make([]coursez.Resource, 0)
	key := "GetTopicResources" + *topicID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
		err = json.Unmarshal([]byte(result), &currentResources)
		if err != nil {
			log.Errorf("Failed to unmarshal redis value: %v", err.Error())
		}
	}

	if len(currentResources) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.resource where topicid='%s' AND is_active=true   ALLOW FILTERING`, *topicID)
		getTopicrRes := func() (resources []coursez.Resource, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return resources, iter.Select(&resources)
		}
		currentResources, err = getTopicrRes()
		if err != nil {
			return nil, err
		}
	}
	gproject := googleprojectlib.GetGoogleProjectID()
	topicsRes := make([]*model.TopicResource, len(currentResources))
	var wg sync.WaitGroup
	for i, topRes := range currentResources {
		mod := topRes
		wg.Add(1)
		go func(i int, mod coursez.Resource) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			url := mod.Url
			if mod.BucketPath != "" {
				storageC := bucket.NewStorageHandler()
				err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
				if err != nil {
					log.Errorf("Failed to initialize storage: %v", err.Error())
					return
				}
				url = storageC.GetSignedURLForObject(mod.BucketPath)
			}
			currentRes := &model.TopicResource{
				ID:        &mod.ID,
				Name:      &mod.Name,
				Type:      &mod.Type,
				URL:       &url,
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
				CreatedBy: &mod.CreatedBy,
				TopicID:   &mod.TopicId,
				CourseID:  &mod.CourseId,
				UpdatedBy: &mod.UpdatedBy,
			}

			topicsRes[i] = currentRes
			wg.Done()
		}(i, mod)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentResources)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topicsRes, nil
}

func GetCourseResources(ctx context.Context, courseID *string) ([]*model.TopicResource, error) {
	currentResources := make([]coursez.Resource, 0)
	key := "GetCourseResources" + *courseID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
		err = json.Unmarshal([]byte(result), &currentResources)
		if err != nil {
			log.Errorf("Failed to unmarshal redis value: %v", err.Error())
		}
	}

	if len(currentResources) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.resource where courseid='%s' AND is_active=true  ALLOW FILTERING`, *courseID)
		getTopicrRes := func() (resources []coursez.Resource, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return resources, iter.Select(&resources)
		}
		currentResources, err = getTopicrRes()
		if err != nil {
			return nil, err
		}
	}
	gproject := googleprojectlib.GetGoogleProjectID()
	topicsRes := make([]*model.TopicResource, len(currentResources))
	var wg sync.WaitGroup
	for i, topRes := range currentResources {
		mod := topRes
		wg.Add(1)
		go func(mod coursez.Resource, i int) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			url := mod.Url
			if mod.BucketPath != "" {
				storageC := bucket.NewStorageHandler()
				err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
				if err != nil {
					log.Errorf("Failed to initialize storage: %v", err.Error())
				}
				url = storageC.GetSignedURLForObject(mod.BucketPath)
			}
			currentRes := &model.TopicResource{
				ID:        &mod.ID,
				Name:      &mod.Name,
				Type:      &mod.Type,
				URL:       &url,
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
				CreatedBy: &mod.CreatedBy,
				TopicID:   &mod.TopicId,
				CourseID:  &mod.CourseId,
				UpdatedBy: &mod.UpdatedBy,
			}

			topicsRes[i] = currentRes
			wg.Done()
		}(mod, i)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentResources)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topicsRes, nil
}
