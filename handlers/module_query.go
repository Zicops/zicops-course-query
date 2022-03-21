package handlers

import (
	"context"
	"fmt"

	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
)

func GetModulesCourseByID(ctx context.Context, courseID *string) ([]*model.Module, error) {
	modules := []*model.Module{}
	qryStr := fmt.Sprintf(`SELECT * from coursez.module where course_id='%s'`, *courseID)
	err := global.CassSession.Session.Query(qryStr, nil).Scan(&modules)
	if err != nil {
		return nil, err
	}
	return modules, nil
}

func GetModuleByID(ctx context.Context, moduleID *string) (*model.Module, error) {
	module := &model.Module{}
	qryStr := fmt.Sprintf(`SELECT * from coursez.module where id='%s'`, *moduleID)
	err := global.CassSession.Session.Query(qryStr, nil).Scan(module)
	if err != nil {
		return nil, err
	}
	return module, nil
}