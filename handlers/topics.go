package handlers

import (
	"context"
	"encoding/base64"
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

func GetTopicsCourseByID(ctx context.Context, courseID *string) ([]*model.Topic, error) {
	currentTopics := make([]coursez.Topic, 0)
	key := "GetTopicsCourseByID" + *courseID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &currentTopics)
		if err != nil {
			log.Errorf("Failed to unmarshal topics: %v", err.Error())
		}
	}
	courses, err := GetCourseByID(ctx, []*string{courseID})
	if err != nil {
		return nil, err
	}
	course := courses[0]

	if len(currentTopics) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.topic where courseid='%s' AND is_active=true  AND lsp_id='%s' ALLOW FILTERING`, *courseID, *course.LspID)
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
	gproject := googleprojectlib.GetGoogleProjectID()
	topicsOut := make([]*model.Topic, len(currentTopics))
	if len(currentTopics) <= 0 {
		return topicsOut, nil
	}
	var wg sync.WaitGroup
	for i, topCopied := range currentTopics {
		mm := topCopied
		wg.Add(1)
		go func(i int, mod coursez.Topic) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			url := ""
			if mod.ImageBucket != "" {
				key := base64.StdEncoding.EncodeToString([]byte(mod.ImageBucket))
				res, err := redis.GetRedisValue(ctx, key)
				if err == nil && res != "" {
					url = res
				} else {
					storageC := bucket.NewStorageHandler()
					err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
					if err != nil {
						log.Errorf("Failed to initialize storage: %v", err.Error())
					}
					url = storageC.GetSignedURLForObject(mod.ImageBucket)
					redis.SetRedisValue(ctx, key, url)
					redis.SetTTL(ctx, key, 3000)
				}
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

			topicsOut[i] = currentModule
			wg.Done()
		}(i, mm)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentTopics)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return topicsOut, nil
}

func GetTopicByID(ctx context.Context, topicID *string) (*model.Topic, error) {
	currentTopics := make([]coursez.Topic, 0)
	key := "GetTopicByID" + *topicID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
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

		qryStr := fmt.Sprintf(`SELECT * from coursez.topic where id='%s' AND is_active=true  ALLOW FILTERING`, *topicID)
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
	gproject := googleprojectlib.GetGoogleProjectID()
	topics := make([]*model.Topic, len(currentTopics))
	if len(currentTopics) <= 0 {
		return nil, nil
	}
	var wg sync.WaitGroup
	for i, copiedTop := range currentTopics {
		tt := copiedTop
		wg.Add(1)
		go func(i int, top coursez.Topic) {
			createdAt := strconv.FormatInt(top.CreatedAt, 10)
			updatedAt := strconv.FormatInt(top.UpdatedAt, 10)
			url := ""
			if top.ImageBucket != "" {
				key := base64.StdEncoding.EncodeToString([]byte(top.ImageBucket))
				res, err := redis.GetRedisValue(ctx, key)
				if err == nil && res != "" {
					url = res
				} else {
					storageC := bucket.NewStorageHandler()
					err = storageC.InitializeStorageClient(ctx, gproject, top.LspId)
					if err != nil {
						log.Errorf("Failed to initialize storage: %v", err.Error())
					}
					url = storageC.GetSignedURLForObject(top.ImageBucket)
					redis.SetRedisValue(ctx, key, url)
					redis.SetTTL(ctx, key, 3000)
				}
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
			topics[i] = currentTop
			wg.Done()
		}(i, tt)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentTopics)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return topics[0], nil
}
