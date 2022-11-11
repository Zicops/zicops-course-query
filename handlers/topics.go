package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	topicsOut := make([]*model.Topic, 0)
	currentTopics := make([]coursez.Topic, 0)
	key := "GetTopicsCourseByID" + *courseID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
		err = json.Unmarshal([]byte(result), &currentTopics)
		if err != nil {
			log.Errorf("Failed to unmarshal topics: %v", err.Error())
		}
	}
	course, err := GetCourseByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

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
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	for _, topCopied := range currentTopics {
		mod := topCopied
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		url := mod.Image
		urlDiff := time.Now().Unix() - mod.UpdatedAt
		needUrl := true
		if urlDiff < 86400 {
			needUrl = false
		}
		if mod.ImageBucket != "" && needUrl {
			err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
				return nil, err
			}
			url = storageC.GetSignedURLForObject(mod.ImageBucket)
			session, err := cassandra.GetCassSession("coursez")
			if err != nil {
				return nil, err
			}
			qryStr := fmt.Sprintf(`UPDATE coursez.topic SET image='%s', updated_at=%d WHERE id='%s' AND is_active=true AND lsp_id='%s'`, url, time.Now().Unix(), mod.ID, mod.LspId)
			err = session.Query(qryStr, nil).Exec()
			if err != nil {
				log.Errorf("Failed to update topic image: %v", err.Error())
				return nil, err
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
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
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
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	for _, copiedTop := range currentTopics {
		top := copiedTop
		createdAt := strconv.FormatInt(top.CreatedAt, 10)
		updatedAt := strconv.FormatInt(top.UpdatedAt, 10)
		url := top.Image
		urlDiff := time.Now().Unix() - top.UpdatedAt
		needUrl := true
		if urlDiff < 86400 {
			needUrl = false
		}
		if top.ImageBucket != "" && needUrl {
			err = storageC.InitializeStorageClient(ctx, gproject, top.LspId)
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
				return nil, err
			}
			url = storageC.GetSignedURLForObject(top.ImageBucket)
			session, err := cassandra.GetCassSession("coursez")
			if err != nil {
				return nil, err
			}
			qryStr := fmt.Sprintf(`UPDATE coursez.topic SET image='%s', updated_at=%d WHERE id='%s' AND is_active=true AND lsp_id='%s'`, url, time.Now().Unix(), top.ID, top.LspId)
			err = session.Query(qryStr, nil).Exec()
			if err != nil {
				log.Errorf("Failed to update topic image: %v", err.Error())
				return nil, err
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
		topics = append(topics, currentTop)
	}
	redisBytes, err := json.Marshal(currentTopics)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topics[0], nil
}
