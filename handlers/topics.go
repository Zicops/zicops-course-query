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
				storageC := bucket.NewStorageHandler()
				err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
				if err != nil {
					log.Errorf("Failed to initialize storage: %v", err.Error())
				}
				url = storageC.GetSignedURLForObjectCache(ctx, mod.ImageBucket)
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
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 60)
	}
	return topicsOut, nil
}

func GetTopicsByCourseIds(ctx context.Context, courseIds []*string, Type *string) ([]*model.Topic, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Println("Got error while getting claims: ", err)
		return nil, err
	}

	var TopicData []coursez.Topic
	key := "GetTopicsCourseByID"
	for _, vv := range courseIds {
		v := *vv
		key = key + v
	}
	role := strings.ToLower(claims["role"].(string))

	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" && result != "" {
		err = json.Unmarshal([]byte(result), &TopicData)
		if err != nil {
			log.Errorf("Failed to unmarshal topics: %v", err.Error())
		}
	}
	lsp := claims["lsp_id"].(string)

	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session
	gproject := googleprojectlib.GetGoogleProjectID()

	//if cache does not hit, then hit the database and store the data
	if len(TopicData) <= 0 {
		var wg sync.WaitGroup
		for _, vvv := range courseIds {

			vv := *vvv
			wg.Add(1)
			go func(v string, lsp_id string) {
				queryStr := fmt.Sprintf(`SELECT * FROM coursez.topic WHERE courseid = '%s' AND is_active=true  AND lsp_id='%s' `, v, lsp_id)
				if Type != nil {
					queryStr = queryStr + fmt.Sprintf(`AND type='%s' `, *Type)
				}
				queryStr = queryStr + "ALLOW FILTERING"
				getTopics := func() (topics []coursez.Topic, err error) {
					q := CassSession.Query(queryStr, nil)
					defer q.Release()
					iter := q.Iter()
					return topics, iter.Select(&topics)
				}
				currentTopics, err := getTopics()
				if err != nil {
					log.Printf("Got error while getting topics: %v", err)
					return
				}
				if len(currentTopics) == 0 {
					return
				}
				TopicData = append(TopicData, currentTopics...)
				wg.Done()

			}(vv, lsp)
		}
		wg.Wait()
	}

	//else directly map data gotten from cache
	res := make([]*model.Topic, len(TopicData))

	for i, kk := range TopicData {
		k := kk
		url := ""
		createdAt := strconv.FormatInt(k.CreatedAt, 10)
		updatedAt := strconv.FormatInt(k.UpdatedAt, 10)
		if k.ImageBucket != "" {
			storageC := bucket.NewStorageHandler()
			err = storageC.InitializeStorageClient(ctx, gproject, k.LspId)
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
			}
			url = storageC.GetSignedURLForObjectCache(ctx, k.ImageBucket)
		}

		currentModule := &model.Topic{
			ID:          &k.ID,
			CourseID:    &k.CourseID,
			ModuleID:    &k.ModuleID,
			ChapterID:   &k.ChapterID,
			Name:        &k.Name,
			Description: &k.Description,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			Sequence:    &k.Sequence,
			CreatedBy:   &k.CreatedBy,
			UpdatedBy:   &k.UpdatedBy,
			Image:       &url,
			Type:        &k.Type,
		}
		res[i] = currentModule
	}

	redisBytes, err := json.Marshal(TopicData)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 60)
	}
	return res, nil
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
				storageC := bucket.NewStorageHandler()
				err = storageC.InitializeStorageClient(ctx, gproject, top.LspId)
				if err != nil {
					log.Errorf("Failed to initialize storage: %v", err.Error())
				}
				url = storageC.GetSignedURLForObjectCache(ctx, top.ImageBucket)
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
