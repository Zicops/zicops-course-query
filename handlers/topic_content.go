package handlers

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetTopicContent(ctx context.Context, topicID *string) ([]*model.TopicContent, error) {
	topicsOut := make([]*model.TopicContent, 0)
	qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where topicid='%s' ALLOW FILTERING`, *topicID)
	getTopicContent := func() (content []coursez.TopicContent, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
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
	for _, topCon := range currentContent {
		mod := topCon
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		urlSub := ""
		if mod.SubtitleFileBucket != "" {
			urlSub = storageC.GetSignedURLForObject(mod.SubtitleFileBucket)
		}
		urlCon := ""
		if mod.TopicContentBucket != "" {
			urlCon = storageC.GetSignedURLForObject(mod.TopicContentBucket)
		}
		currentModule := &model.TopicContent{
			ID:                &mod.ID,
			Language:          &mod.Language,
			TopicID:           &mod.TopicId,
			SubtitleURL:       &urlSub,
			ContentURL:        &urlCon,
			CreatedAt:         &createdAt,
			UpdatedAt:         &updatedAt,
			StartTime:         &mod.StartTime,
			Duration:          &mod.Duration,
			SkipIntroDuration: &mod.SkipIntroDuration,
			NextShowTime:      &mod.NextShowtime,
			FromEndTime:       &mod.FromEndTime,
			Type:              &mod.Type,
		}

		topicsOut = append(topicsOut, currentModule)
	}
	return topicsOut, nil
}

func GetTopicContentByCourse(ctx context.Context, courseID *string) ([]*model.TopicContent, error) {
	topicsOut := make([]*model.TopicContent, 0)
	qryStr := fmt.Sprintf(`SELECT * from coursez.topic_content where courseid='%s' ALLOW FILTERING`, *courseID)
	getTopicContent := func() (content []coursez.TopicContent, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
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
	for _, topCon := range currentContent {
		mod := topCon
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		urlSub := ""
		if mod.SubtitleFileBucket != "" {
			urlSub = storageC.GetSignedURLForObject(mod.SubtitleFileBucket)
		}
		urlCon := ""
		if mod.TopicContentBucket != "" {
			urlCon = storageC.GetSignedURLForObject(mod.TopicContentBucket)
		}
		currentModule := &model.TopicContent{
			ID:                &mod.ID,
			Language:          &mod.Language,
			TopicID:           &mod.TopicId,
			SubtitleURL:       &urlSub,
			ContentURL:        &urlCon,
			CreatedAt:         &createdAt,
			UpdatedAt:         &updatedAt,
			StartTime:         &mod.StartTime,
			Duration:          &mod.Duration,
			SkipIntroDuration: &mod.SkipIntroDuration,
			NextShowTime:      &mod.NextShowtime,
			FromEndTime:       &mod.FromEndTime,
			Type:              &mod.Type,
		}

		topicsOut = append(topicsOut, currentModule)
	}
	return topicsOut, nil
}
