package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
)

func GetModulesCourseByID(ctx context.Context, courseID *string) ([]*model.Module, error) {
	modules := make([]*model.Module, 0)
	key := "GetModulesCourseByID" + *courseID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &modules)
		if err == nil {
			return modules, nil
		}
	}

	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session
	courses, err := GetCourseByID(ctx, []*string{courseID})
	if err != nil {
		return nil, err
	}
	course := courses[0]
	lspId := course.LspID

	qryStr := fmt.Sprintf(`SELECT * from coursez.module where courseid='%s' AND lsp_id='%s' AND is_active=true ALLOW FILTERING`, *courseID, *lspId)
	getModules := func() (modules []coursez.Module, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return modules, iter.Select(&modules)
	}
	currentModules, err := getModules()
	if err != nil {
		return nil, err
	}
	if len(currentModules) <= 0 {
		return modules, nil
	}
	modules = make([]*model.Module, len(currentModules))
	var wg sync.WaitGroup
	for i, copiedMod := range currentModules {
		cm := copiedMod
		wg.Add(1)
		go func(i int, mod coursez.Module) {
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
			modules[i] = currentModule
			wg.Done()
		}(i, cm)
	}
	wg.Wait()
	redisBtres, err := json.Marshal(modules)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBtres))
	}
	return modules, nil
}

func GetModuleByID(ctx context.Context, moduleID *string) (*model.Module, error) {
	modules := make([]*model.Module, 0)
	key := "GetModuleByID" + *moduleID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &modules)
		if err == nil {
			return modules[0], nil
		}
	}

	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from coursez.module where id='%s' AND is_active=true ALLOW FILTERING`, *moduleID)
	getModules := func() (modules []coursez.Module, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return modules, iter.Select(&modules)
	}
	currentModules, err := getModules()
	if err != nil {
		return nil, err
	}
	modules = make([]*model.Module, len(currentModules))
	if len(currentModules) <= 0 {
		return nil, nil
	}
	var wg sync.WaitGroup
	for i, copiedMod := range currentModules {
		cm := copiedMod
		wg.Add(1)
		go func(i int, mod coursez.Module) {
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
			modules[i] = currentModule
			wg.Done()
		}(i, cm)
	}
	wg.Wait()
	redisBtres, err := json.Marshal(modules)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBtres))
	}

	return modules[0], nil
}
