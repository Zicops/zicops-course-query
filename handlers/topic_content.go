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
	"github.com/zicops/zicops-course-query/constants"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetTopicContent(ctx context.Context, topicID *string) ([]*model.TopicContent, error) {
	currentContent := make([]coursez.TopicContent, 0)
	key := "GetTopicContent" + *topicID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &currentContent)
		if err != nil {
			log.Errorf("Error in unmarshalling redis value for key %s", key)
		}
	}
	if len(currentContent) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where topicid='%s' AND is_active=true  ALLOW FILTERING`, *topicID)
		getTopicContent := func() (content []coursez.TopicContent, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return content, iter.Select(&content)
		}
		currentContent, err = getTopicContent()
		if err != nil {
			return nil, err
		}
	}
	gproject := googleprojectlib.GetGoogleProjectID()
	topicsOut := make([]*model.TopicContent, len(currentContent))
	if len(currentContent) <= 0 {
		return topicsOut, nil
	}
	urlSub := make([]*model.SubtitleURL, 0)
	var wg sync.WaitGroup
	for i, topCon := range currentContent {
		mm := topCon
		wg.Add(1)
		go func(i int, mod coursez.TopicContent) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			mainBucket := mod.CourseId + "/" + mod.TopicId + "/subtitles/"
			storageC := bucket.NewStorageHandler()
			err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
			}
			if mainBucket != "" {
				urlSub = storageC.GetSignedURLsForObjects(mainBucket)
			}

			urlCon := mod.Url
			_, ok := constants.StaticTypeMap[mod.Type]
			if mod.TopicContentBucket != "" && !ok {
				urlCon = storageC.GetSignedURLForObject(mod.TopicContentBucket)
			} else if mod.TopicContentBucket != "" && ok {
				urlCon = mod.Url
			}

			currentModule := &model.TopicContent{
				ID:                &mod.ID,
				Language:          &mod.Language,
				TopicID:           &mod.TopicId,
				CourseID:          &mod.CourseId,
				SubtitleURL:       urlSub,
				ContentURL:        &urlCon,
				CreatedAt:         &createdAt,
				UpdatedAt:         &updatedAt,
				StartTime:         &mod.StartTime,
				Duration:          &mod.Duration,
				SkipIntroDuration: &mod.SkipIntroDuration,
				NextShowTime:      &mod.NextShowtime,
				FromEndTime:       &mod.FromEndTime,
				Type:              &mod.Type,
				IsDefault:         &mod.IsDefault,
			}

			topicsOut[i] = currentModule
			wg.Done()
		}(i, mm)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentContent)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return topicsOut, nil
}

func GetTopicExams(ctx context.Context, topicID *string) ([]*model.TopicExam, error) {
	topicsOut := make([]*model.TopicExam, 0)
	key := "GetTopicExams" + *topicID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &topicsOut)
		if err == nil {
			return topicsOut, nil
		}
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from coursez.topic_exam where topicid='%s' AND is_active=true  ALLOW FILTERING`, *topicID)
	getTopicContent := func() (content []coursez.TopicExam, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return content, iter.Select(&content)
	}
	currentContent, err := getTopicContent()
	if err != nil {
		return nil, err
	}
	topicsOut = make([]*model.TopicExam, len(currentContent))
	if len(currentContent) <= 0 {
		return topicsOut, nil
	}
	var wg sync.WaitGroup
	for i, topCon := range currentContent {
		mm := topCon
		wg.Add(1)
		go func(i int, mod coursez.TopicExam) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			currentModule := &model.TopicExam{
				ID:        &mod.ID,
				Language:  &mod.Language,
				TopicID:   &mod.TopicId,
				CourseID:  &mod.CourseId,
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
				ExamID:    &mod.ExamId,
			}

			topicsOut[i] = currentModule
			wg.Done()
		}(i, mm)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(topicsOut)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return topicsOut, nil
}

func GetTopicContentByCourse(ctx context.Context, courseID *string) ([]*model.TopicContent, error) {
	currentContent := make([]coursez.TopicContent, 0)
	key := "GetTopicContentByCourse" + *courseID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &currentContent)
		if err != nil {
			log.Errorf("Error in unmarshalling redis value for key %s", key)
		}
	}
	courses, err := GetCourseByID(ctx, []*string{courseID})
	if err != nil {
		return nil, err
	}
	course := courses[0]
	if len(currentContent) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where courseid='%s' AND is_active=true  AND lsp_id ='%s' ALLOW FILTERING`, *courseID, *course.LspID)
		getTopicContent := func() (content []coursez.TopicContent, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return content, iter.Select(&content)
		}
		currentContent, err = getTopicContent()
		if err != nil {
			return nil, err
		}
	}
	gproject := googleprojectlib.GetGoogleProjectID()
	urlSub := make([]*model.SubtitleURL, 0)
	topicsOut := make([]*model.TopicContent, len(currentContent))
	if len(currentContent) <= 0 {
		return topicsOut, nil
	}
	var wg sync.WaitGroup
	for i, topCon := range currentContent {
		mm := topCon
		wg.Add(1)
		go func(i int, mod coursez.TopicContent) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			storageC := bucket.NewStorageHandler()
			err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
			}
			mainBucket := mod.CourseId + "/" + mod.TopicId + "/subtitles/"
			if mainBucket != "" {
				urlSub = storageC.GetSignedURLsForObjects(mainBucket)
			}

			urlCon := mod.Url
			typeCon := strings.ToLower(mod.Type)
			if mod.TopicContentBucket != "" && (strings.Contains(typeCon, "static") || strings.Contains(typeCon, "scorm") || strings.Contains(typeCon, "tincan") || strings.Contains(typeCon, "cmi5") || strings.Contains(typeCon, "html5")) {
				urlCon = mod.Url
			} else if mod.TopicContentBucket != "" {
				urlCon = storageC.GetSignedURLForObject(mod.TopicContentBucket)
			}
			currentModule := &model.TopicContent{
				ID:                &mod.ID,
				Language:          &mod.Language,
				TopicID:           &mod.TopicId,
				CourseID:          &mod.CourseId,
				SubtitleURL:       urlSub,
				ContentURL:        &urlCon,
				CreatedAt:         &createdAt,
				UpdatedAt:         &updatedAt,
				StartTime:         &mod.StartTime,
				Duration:          &mod.Duration,
				SkipIntroDuration: &mod.SkipIntroDuration,
				NextShowTime:      &mod.NextShowtime,
				FromEndTime:       &mod.FromEndTime,
				Type:              &mod.Type,
				IsDefault:         &mod.IsDefault,
			}

			topicsOut[i] = currentModule
			wg.Done()
		}(i, mm)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentContent)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return topicsOut, nil
}

func GetTopicExamsByCourse(ctx context.Context, courseID *string) ([]*model.TopicExam, error) {
	topicsOut := make([]*model.TopicExam, 0)
	key := "GetTopicExamsByCourse" + *courseID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &topicsOut)
		if err == nil {
			return topicsOut, nil
		}
	}

	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	courses, err := GetCourseByID(ctx, []*string{courseID})
	if err != nil {
		return nil, err
	}
	course := courses[0]
	qryStr := fmt.Sprintf(`SELECT * from coursez.topic_exam where courseid='%s' AND is_active=true  AND lsp_id='%s' ALLOW FILTERING`, *courseID, *course.LspID)
	getTopicContent := func() (content []coursez.TopicExam, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return content, iter.Select(&content)
	}
	currentContent, err := getTopicContent()
	if err != nil {
		return nil, err
	}
	topicsOut = make([]*model.TopicExam, len(currentContent))
	if len(currentContent) <= 0 {
		return topicsOut, nil
	}
	var wg sync.WaitGroup
	for i, topCon := range currentContent {
		mm := topCon
		wg.Add(1)
		go func(i int, mod coursez.TopicExam) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			currentModule := &model.TopicExam{
				ID:        &mod.ID,
				Language:  &mod.Language,
				TopicID:   &mod.TopicId,
				CourseID:  &mod.CourseId,
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
				ExamID:    &mod.ExamId,
			}

			topicsOut[i] = currentModule
			wg.Done()
		}(i, mm)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(topicsOut)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return topicsOut, nil
}

func GetTopicContentByModule(ctx context.Context, moduleID *string) ([]*model.TopicContent, error) {
	currentContent := make([]coursez.TopicContent, 0)
	key := "GetTopicContentByModule" + *moduleID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	/*
		lspID := claims["lsp_id"].(string)
	*/
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &currentContent)
		if err != nil {
			log.Errorf("Error in unmarshalling redis value for key %s", key)
		}
	}
	if len(currentContent) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		//qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where moduleid='%s' AND is_active=true  AND lsp_id ='%s' ALLOW FILTERING`, *moduleID, lspID)
		qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where moduleid='%s' AND is_active=true ALLOW FILTERING`, *moduleID)
		getTopicContent := func() (content []coursez.TopicContent, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return content, iter.Select(&content)
		}
		currentContent, err = getTopicContent()
		if err != nil {
			return nil, err
		}
	}
	gproject := googleprojectlib.GetGoogleProjectID()
	topicsOut := make([]*model.TopicContent, len(currentContent))
	if len(currentContent) <= 0 {
		return topicsOut, nil
	}
	var wg sync.WaitGroup
	for i, topCon := range currentContent {
		mm := topCon
		wg.Add(1)
		go func(i int, mod coursez.TopicContent) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			storageC := bucket.NewStorageHandler()
			err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
			}
			mainBucket := mod.CourseId + "/" + mod.TopicId + "/subtitles/"
			urlSub := make([]*model.SubtitleURL, 0)
			if mainBucket != "" {
				urlSub = storageC.GetSignedURLsForObjects(mainBucket)
			}

			urlCon := mod.Url
			typeCon := strings.ToLower(mod.Type)
			if mod.TopicContentBucket != "" && (strings.Contains(typeCon, "static") || strings.Contains(typeCon, "scorm") || strings.Contains(typeCon, "tincan") || strings.Contains(typeCon, "cmi5") || strings.Contains(typeCon, "html5")) {
				urlCon = mod.Url
			} else if mod.TopicContentBucket != "" {
				urlCon = storageC.GetSignedURLForObject(mod.TopicContentBucket)
			}
			currentModule := &model.TopicContent{
				ID:                &mod.ID,
				Language:          &mod.Language,
				TopicID:           &mod.TopicId,
				CourseID:          &mod.CourseId,
				SubtitleURL:       urlSub,
				ContentURL:        &urlCon,
				CreatedAt:         &createdAt,
				UpdatedAt:         &updatedAt,
				StartTime:         &mod.StartTime,
				Duration:          &mod.Duration,
				SkipIntroDuration: &mod.SkipIntroDuration,
				NextShowTime:      &mod.NextShowtime,
				FromEndTime:       &mod.FromEndTime,
				Type:              &mod.Type,
				IsDefault:         &mod.IsDefault,
			}

			topicsOut[i] = currentModule
			wg.Done()
		}(i, mm)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentContent)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return topicsOut, nil
}
