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

func GetModulesCourseByID(ctx context.Context, courseID *string) ([]*model.Module, error) {
	modules := make([]*model.Module, 0)
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassSession = session
	defer global.CassSession.Close()
	qryStr := fmt.Sprintf(`SELECT * from coursez.module where courseid='%s' ALLOW FILTERING`, *courseID)
	getModules := func() (modules []coursez.Module, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return modules, iter.Select(&modules)
	}
	currentModules, err := getModules()
	if err != nil {
		return nil, err
	}
	for _, copiedMod := range currentModules {
		mod := copiedMod
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		currentModule := &model.Module{
			ID:          &mod.ID,
			CourseID:    &mod.CourseID,
			IsChapter:   &mod.IsChapter,
			Name:        &mod.Name,
			Description: &mod.Description,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			Level:       &mod.Level,
			Owner:       &mod.Owner,
			Sequence:    &mod.Sequence,
			SetGlobal:   &mod.SetGlobal,
			Duration:    &mod.Duration,
		}
		modules = append(modules, currentModule)
	}
	return modules, nil
}

func GetModuleByID(ctx context.Context, moduleID *string) (*model.Module, error) {
	modules := make([]*model.Module, 0)
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassSession = session
	defer global.CassSession.Close()
	qryStr := fmt.Sprintf(`SELECT * from coursez.module where id='%s' ALLOW FILTERING`, *moduleID)
	getModules := func() (modules []coursez.Module, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return modules, iter.Select(&modules)
	}
	currentModules, err := getModules()
	if err != nil {
		return nil, err
	}
	for _, copiedMod := range currentModules {
		mod := copiedMod
		createdAt := strconv.FormatInt(mod.CreatedAt, 10)
		updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
		currentModule := &model.Module{
			ID:          &mod.ID,
			CourseID:    &mod.CourseID,
			IsChapter:   &mod.IsChapter,
			Name:        &mod.Name,
			Description: &mod.Description,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			Level:       &mod.Level,
			Owner:       &mod.Owner,
			Sequence:    &mod.Sequence,
			SetGlobal:   &mod.SetGlobal,
			Duration:    &mod.Duration,
		}
		modules = append(modules, currentModule)
	}
	return modules[0], nil
}
