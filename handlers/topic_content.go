package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetTopicContent(ctx context.Context, topicID *string) ([]*model.TopicContent, error) {
	topicsOut := make([]*model.TopicContent, 0)
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassSession = session

	qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where topicid='%s' ALLOW FILTERING`, *topicID)
	getTopicContent := func() (content []coursez.TopicContent, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return content, iter.Select(&content)
	}
	currentContent, err := getTopicContent()
	if err != nil {
		return nil, err
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to initialize storage: %v", err.Error())
		return nil, err
	}

	urlSub := make([]*model.SubtitleURL, 0)
	for _, topCon := range currentContent {
		mod := topCon
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		mainBucket := mod.CourseId + "/" + mod.TopicId + "/subtitles/"
		if mainBucket != "" {
			urlSub = storageC.GetSignedURLsForObjects(mainBucket)
		}

		urlCon := ""
		if mod.TopicContentBucket != "" {
			urlCon = storageC.GetSignedURLForObject(mod.TopicContentBucket)
		} else {
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
	return topicsOut, nil
}

func GetTopicExams(ctx context.Context, topicID *string) ([]*model.TopicExam, error) {
	topicsOut := make([]*model.TopicExam, 0)
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassSession = session

	qryStr := fmt.Sprintf(`SELECT * from coursez.topic_exam where topicid='%s' ALLOW FILTERING`, *topicID)
	getTopicContent := func() (content []coursez.TopicExam, err error) {
		q := global.CassSession.Query(qryStr, nil)
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
	return topicsOut, nil
}

func GetTopicContentByCourse(ctx context.Context, courseID *string) ([]*model.TopicContent, error) {
	topicsOut := make([]*model.TopicContent, 0)
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassSession = session

	qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where courseid='%s' ALLOW FILTERING`, *courseID)
	getTopicContent := func() (content []coursez.TopicContent, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return content, iter.Select(&content)
	}
	currentContent, err := getTopicContent()
	if err != nil {
		return nil, err
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to initialize storage: %v", err.Error())
		return nil, err
	}
	urlSub := make([]*model.SubtitleURL, 0)
	for _, topCon := range currentContent {
		mod := topCon
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)

		mainBucket := mod.CourseId + "/" + mod.TopicId + "/subtitles/"
		if mainBucket != "" {
			urlSub = storageC.GetSignedURLsForObjects(mainBucket)
		}

		urlCon := ""
		typeCon := strings.ToLower(mod.Type)
		if mod.TopicContentBucket != "" && (strings.Contains(typeCon, "static") || strings.Contains(typeCon, "scorm") || strings.Contains(typeCon, "tincan") || strings.Contains(typeCon, "cmi5") || strings.Contains(typeCon, "html5")) {
			urlCon = storageC.GetSignedURLForObjectPub(mod.TopicContentBucket)
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

		topicsOut = append(topicsOut, currentModule)
	}

	return topicsOut, nil
}

func GetTopicExamsByCourse(ctx context.Context, courseID *string) ([]*model.TopicExam, error) {
	topicsOut := make([]*model.TopicExam, 0)
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassSession = session

	qryStr := fmt.Sprintf(`SELECT * from coursez.topic_exam where courseid='%s' ALLOW FILTERING`, *courseID)
	getTopicContent := func() (content []coursez.TopicExam, err error) {
		q := global.CassSession.Query(qryStr, nil)
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

	return topicsOut, nil
}
