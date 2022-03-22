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
	qryStr := fmt.Sprintf(`SELECT * from coursez.topic where course_id='%s' ALLOW FILTERING`, *courseID)
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
		}

		topicsOut = append(topicsOut, currentModule)
	}
	return topicsOut, nil
}
