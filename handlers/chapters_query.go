package handlers

import (
	"context"
	"fmt"

	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
)

func GetChaptersCourseByID(ctx context.Context, courseID *string) ([]*model.Chapter, error) {
	chapters := []*model.Chapter{}
	qryStr := fmt.Sprintf(`SELECT * from coursez.chapter where course_id='%s'`, *courseID)
	err := global.CassSession.Session.Query(qryStr, nil).Scan(&chapters)
	if err != nil {
		return nil, err
	}
	return chapters, nil
}

func GetChapterByID(ctx context.Context, chapterID *string) (*model.Chapter, error) {
	chapter := &model.Chapter{}
	qryStr := fmt.Sprintf(`SELECT * from coursez.chapter where id='%s'`, *chapterID)
	err := global.CassSession.Session.Query(qryStr, nil).Scan(chapter)
	if err != nil {
		return nil, err
	}
	return chapter, nil
}
