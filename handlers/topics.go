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

func GetTopicsCourseByID(ctx context.Context, courseID *string) ([]*model.Topic, error) {
	topicsOut := make([]*model.Topic, 0)
	qryStr := fmt.Sprintf(`SELECT * from coursez.topic where courseid='%s' ALLOW FILTERING`, *courseID)
	getTopics := func() (topics []coursez.Topic, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return topics, iter.Select(&topics)
	}
	currentTopics, err := getTopics()
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
	for _, topCopied := range currentTopics {
		mod := topCopied
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		url := storageC.GetSignedURLForObject(mod.Image)
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
	return topicsOut, nil
}

func GetTopicByID(ctx context.Context, topicID *string) (*model.Topic, error) {
	topics := make([]*model.Topic, 0)
	qryStr := fmt.Sprintf(`SELECT * from coursez.topic where id='%s' ALLOW FILTERING`, *topicID)
	getTopics := func() (topics []coursez.Topic, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return topics, iter.Select(&topics)
	}
	currentTopics, err := getTopics()
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
	for _, copiedTop := range currentTopics {
		top := copiedTop
		createdAt := strconv.FormatInt(top.CreatedAt, 10)
		updatedAt := strconv.FormatInt(top.UpdatedAt, 10)
		url := storageC.GetSignedURLForObject(top.Image)
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
	return topics[0], nil
}
