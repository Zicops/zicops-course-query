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

func GetTopicsCourseByID(ctx context.Context, courseID *string) ([]*model.Topic, error) {
	topicsOut := make([]*model.Topic, 0)
	currentTopics := make([]coursez.Topic, 0)
	key := "GetTopicsCourseByID" + *courseID
	result, err := redis.GetRedisValue(key)
	if err == nil {
		err = json.Unmarshal([]byte(result), &currentTopics)
		if err != nil {
			log.Errorf("Failed to unmarshal topics: %v", err.Error())
		}
	}
	if len(currentTopics) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.topic where courseid='%s' ALLOW FILTERING`, *courseID)
		getTopics := func() (topics []coursez.Topic, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return topics, iter.Select(&topics)
		}
		currentTopics, err = getTopics()
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
	for _, topCopied := range currentTopics {
		mod := topCopied
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		url := ""
		if mod.ImageBucket != "" {
			url = storageC.GetSignedURLForObject(mod.ImageBucket)
		}
		currentModule := &model.Topic{
			ID:          &mod.ID,
			CourseID:    &mod.CourseID,
			ModuleID:    &mod.ModuleID,
			ChapterID:   &mod.ChapterID,
			Name:        &mod.Name,
			Description: &mod.Description,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			Sequence:    &mod.Sequence,
			CreatedBy:   &mod.CreatedBy,
			UpdatedBy:   &mod.UpdatedBy,
			Image:       &url,
			Type:        &mod.Type,
		}

		topicsOut = append(topicsOut, currentModule)
	}
	redisBytes, err := json.Marshal(currentTopics)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topicsOut, nil
}

func GetTopicByID(ctx context.Context, topicID *string) (*model.Topic, error) {
	topics := make([]*model.Topic, 0)
	currentTopics := make([]coursez.Topic, 0)
	key := "GetTopicByID" + *topicID
	result, err := redis.GetRedisValue(key)
	if err == nil {
		err = json.Unmarshal([]byte(result), &currentTopics)
		if err != nil {
			log.Errorf("Failed to unmarshal topics: %v", err.Error())
		}
	}
	if len(currentTopics) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.topic where id='%s' ALLOW FILTERING`, *topicID)
		getTopics := func() (topics []coursez.Topic, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return topics, iter.Select(&topics)
		}
		currentTopics, err = getTopics()
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
	for _, copiedTop := range currentTopics {
		top := copiedTop
		createdAt := strconv.FormatInt(top.CreatedAt, 10)
		updatedAt := strconv.FormatInt(top.UpdatedAt, 10)
		url := ""
		if top.ImageBucket != "" {
			url = storageC.GetSignedURLForObject(top.ImageBucket)
		}
		currentTop := &model.Topic{
			ID:          &top.ID,
			CourseID:    &top.CourseID,
			ModuleID:    &top.ModuleID,
			ChapterID:   &top.ChapterID,
			Name:        &top.Name,
			Description: &top.Description,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			Sequence:    &top.Sequence,
			CreatedBy:   &top.CreatedBy,
			UpdatedBy:   &top.UpdatedBy,
			Image:       &url,
			Type:        &top.Type,
		}
		topics = append(topics, currentTop)
	}
	redisBytes, err := json.Marshal(currentTopics)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topics[0], nil
}
