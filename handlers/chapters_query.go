package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
)

func GetChaptersCourseByID(ctx context.Context, courseID *string) ([]*model.Chapter, error) {
	chapters := make([]*model.Chapter, 0)
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassSession = session
	defer global.CassSession.Close()
	qryStr := fmt.Sprintf(`SELECT * from coursez.chapter where courseid='%s' ALLOW FILTERING`, *courseID)
	getChapters := func() (modules []coursez.Chapter, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return modules, iter.Select(&modules)
	}
	currentChapters, err := getChapters()
	if err != nil {
		return nil, err
	}
	for _, copiedMod := range currentChapters {
		mod := copiedMod
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		currentChapter := &model.Chapter{
			ID:          &mod.ID,
			CourseID:    &mod.CourseID,
			Description: &mod.Description,
			ModuleID:    &mod.ModuleID,
			Name:        &mod.Name,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			Sequence:    &mod.Sequence,
		}
		chapters = append(chapters, currentChapter)
	}
	return chapters, nil
}

func GetChapterByID(ctx context.Context, chapterID *string) (*model.Chapter, error) {
	chapters := make([]*model.Chapter, 0)
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassSession = session
	defer global.CassSession.Close()
	qryStr := fmt.Sprintf(`SELECT * from coursez.chapter where id='%s' ALLOW FILTERING`, *chapterID)
	getChapters := func() (modules []coursez.Chapter, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return modules, iter.Select(&modules)
	}
	currentChapters, err := getChapters()
	if err != nil {
		return nil, err
	}
	for _, copiedMod := range currentChapters {
		mod := copiedMod
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		currentChapter := &model.Chapter{
			ID:          &mod.ID,
			CourseID:    &mod.CourseID,
			Description: &mod.Description,
			ModuleID:    &mod.ModuleID,
			Name:        &mod.Name,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			Sequence:    &mod.Sequence,
		}
		chapters = append(chapters, currentChapter)
	}
	return chapters[0], nil
}
