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
	"github.com/zicops/zicops-course-query/constants"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetTopicContent(ctx context.Context, topicID *string) ([]*model.TopicContent, error) {
	topicsOut := make([]*model.TopicContent, 0)
	currentContent := make([]coursez.TopicContent, 0)
	key := "GetTopicContent" + *topicID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
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
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()

	urlSub := make([]*model.SubtitleURL, 0)
	for _, topCon := range currentContent {
		mod := topCon
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		mainBucket := mod.CourseId + "/" + mod.TopicId + "/subtitles/"
		err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
		if err != nil {
			log.Errorf("Failed to initialize storage: %v", err.Error())
			return nil, err
		}
		if mainBucket != "" {
			urlSub = storageC.GetSignedURLsForObjects(mainBucket)
		}

		urlCon := mod.Url
		_, ok := constants.StaticTypeMap[mod.Type]
		urlDiff := time.Now().Unix() - mod.UpdatedAt
		needUrl := true
		if urlDiff < 604800 {
			needUrl = false
		}
		if mod.TopicContentBucket != "" && !ok && needUrl {
			urlCon = storageC.GetSignedURLForObject(mod.TopicContentBucket)
			session, err := cassandra.GetCassSession("coursez")
			if err != nil {
				return nil, err
			}
			CassSession := session
			qryStr := fmt.Sprintf(`UPDATE coursez.topic_content SET url='%s', updated_at=%d WHERE id='%s' AND lsp_id='%s' AND is_active=true`, urlCon, time.Now().Unix(), mod.ID, mod.LspId)
			updateTopicContent := func() error {
				q := CassSession.Query(qryStr, nil)
				defer q.Release()
				return q.Exec()
			}
			err = updateTopicContent()
			if err != nil {
				return nil, err
			}
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

		topicsOut = append(topicsOut, currentModule)
	}
	redisBytes, err := json.Marshal(currentContent)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
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
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
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
	for _, topCon := range currentContent {
		mod := topCon
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

		topicsOut = append(topicsOut, currentModule)
	}
	redisBytes, err := json.Marshal(topicsOut)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
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
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
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

	course, err := GetCourseByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
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

	for _, topCon := range currentContent {
		mod := topCon
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

		topicsOut = append(topicsOut, currentModule)
	}
	redisBytes, err := json.Marshal(topicsOut)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topicsOut, nil
}

func GetTopicContentByModule(ctx context.Context, moduleID *string) ([]*model.TopicContent, error) {
	topicsOut := make([]*model.TopicContent, 0)
	currentContent := make([]coursez.TopicContent, 0)
	key := "GetTopicContentByModule" + *moduleID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	lspID := claims["lsp_id"].(string)
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
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

		qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where moduleid='%s' AND is_active=true  AND lsp_id ='%s' ALLOW FILTERING`, *moduleID, lspID)
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
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	urlSub := make([]*model.SubtitleURL, 0)
	for _, topCon := range currentContent {
		mod := topCon
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
		if err != nil {
			log.Errorf("Failed to initialize storage: %v", err.Error())
			return nil, err
		}
		mainBucket := mod.CourseId + "/" + mod.TopicId + "/subtitles/"
		if mainBucket != "" {
			urlSub = storageC.GetSignedURLsForObjects(mainBucket)
		}
		urlDiff := time.Now().Unix() - mod.UpdatedAt
		needUrl := true
		if urlDiff < 604800 {
			needUrl = false
		}
		urlCon := mod.Url
		typeCon := strings.ToLower(mod.Type)
		if mod.TopicContentBucket != "" && (strings.Contains(typeCon, "static") || strings.Contains(typeCon, "scorm") || strings.Contains(typeCon, "tincan") || strings.Contains(typeCon, "cmi5") || strings.Contains(typeCon, "html5")) {
			urlCon = mod.Url
		} else if mod.TopicContentBucket != "" && needUrl {
			urlCon = storageC.GetSignedURLForObject(mod.TopicContentBucket)
			session, err := cassandra.GetCassSession("coursez")
			if err != nil {
				return nil, err
			}
			CassSession := session
			qryStr := fmt.Sprintf(`UPDATE coursez.topic_content SET url='%s', updated_at=%d WHERE id='%s' AND lsp_id='%s' AND is_active=true`, urlCon, time.Now().Unix(), mod.ID, mod.LspId)
			updateTopicContent := func() (err error) {
				q := CassSession.Query(qryStr, nil)
				defer q.Release()
				return q.Exec()
			}
			err = updateTopicContent()
			if err != nil {
				return nil, err
			}
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

		topicsOut = append(topicsOut, currentModule)
	}
	redisBytes, err := json.Marshal(currentContent)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topicsOut, nil
}
