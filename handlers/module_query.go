package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
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

	qryStr := fmt.Sprintf(`SELECT * from coursez.module where courseid='%s' ALLOW FILTERING`, *courseID)
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
	redisBtres, err := json.Marshal(modules)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBtres))
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
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin"{
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

	qryStr := fmt.Sprintf(`SELECT * from coursez.module where id='%s' ALLOW FILTERING`, *moduleID)
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
	redisBtres, err := json.Marshal(modules)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBtres))
	}

	return modules[0], nil
}
