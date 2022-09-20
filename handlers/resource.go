package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetTopicResources(ctx context.Context, topicID *string) ([]*model.TopicResource, error) {
	topicsRes := make([]*model.TopicResource, 0)
	currentResources := make([]coursez.Resource, 0)
	key := "GetTopicResources" + *topicID
	result, err := redis.GetRedisValue(key)
	if err == nil {
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

		qryStr := fmt.Sprintf(`SELECT * from coursez.resource where topicid='%s' ALLOW FILTERING`, *topicID)
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
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to initialize storage: %v", err.Error())
		return nil, err
	}
	for _, topRes := range currentResources {
		mod := topRes
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		url := mod.Url
		if mod.BucketPath != "" {
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

		topicsRes = append(topicsRes, currentRes)
	}
	redisBytes, err := json.Marshal(currentResources)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topicsRes, nil
}

func GetCourseResources(ctx context.Context, courseID *string) ([]*model.TopicResource, error) {
	topicsRes := make([]*model.TopicResource, 0)
	currentResources := make([]coursez.Resource, 0)
	key := "GetCourseResources" + *courseID
	result, err := redis.GetRedisValue(key)
	if err == nil {
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

		qryStr := fmt.Sprintf(`SELECT * from coursez.resource where courseid='%s' ALLOW FILTERING`, *courseID)
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
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to initialize storage: %v", err.Error())
		return nil, err
	}
	for _, topRes := range currentResources {
		mod := topRes
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		url := mod.Url
		if mod.BucketPath != "" {
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

		topicsRes = append(topicsRes, currentRes)
	}

	redisBytes, err := json.Marshal(currentResources)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topicsRes, nil
}
