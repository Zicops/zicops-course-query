package handlers

import (
	"context"
	"fmt"

	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
)

func GetCourseByID(ctx context.Context, courseID *string) (*model.Course, error) {
	course := &model.Course{}
	qryStr := fmt.Sprintf(`SELECT * from coursez.course where id='%s'`, *courseID)
	err := global.CassSession.Session.Query(qryStr, nil).Scan(course)
	if err != nil {
		return nil, err
	}
	return course, nil
}
