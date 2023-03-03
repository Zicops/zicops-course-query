package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
)

func GetChaptersCourseByID(ctx context.Context, courseID *string) ([]*model.Chapter, error) {
	chapters := make([]*model.Chapter, 0)
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	key := "GetChaptersCourseByID" + *courseID
	result, err := redis.GetRedisValue(ctx, key)

	if err != nil {
		log.Errorf("GetChaptersCourseByID: %v", err)
	}
	if result != "" {
		err = json.Unmarshal([]byte(result), &chapters)
		if err != nil {
			log.Errorf("GetChaptersCourseByID: %v", err)
		}
	}
	if len(chapters) > 0 && role == "learner" {
		return chapters, nil
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session
	qryStr := fmt.Sprintf(`SELECT * from coursez.chapter where courseid='%s' ALLOW FILTERING`, *courseID)
	getChapters := func() (modules []coursez.Chapter, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return modules, iter.Select(&modules)
	}
	currentChapters, err := getChapters()
	if err != nil {
		return nil, err
	}
	chapters = make([]*model.Chapter, len(currentChapters))
	if len(currentChapters) <= 0 {
		return chapters, nil
	}
	var wg sync.WaitGroup
	for i, copiedMod := range currentChapters {
		mm := copiedMod
		wg.Add(1)
		go func(i int, mod coursez.Chapter) {
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
			chapters[i] = currentChapter
			wg.Done()
		}(i, mm)
	}
	wg.Wait()
	chaptersBytes, err := json.Marshal(chapters)
	if err != nil {
		log.Errorf("GetChaptersCourseByID: %v", err)
	} else {
		err = redis.SetRedisValue(ctx, key, string(chaptersBytes))
		if err != nil {
			log.Errorf("GetChaptersCourseByID: %v", err)
		}
		redis.SetTTL(ctx, key, 60)
	}
	return chapters, nil
}

func GetChapterByID(ctx context.Context, chapterID *string) (*model.Chapter, error) {
	chapters := make([]*model.Chapter, 0)
	key := "GetChapterByID" + *chapterID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err != nil {
		log.Errorf("GetChapterByID: %v", err)
	}
	if result != "" {
		err = json.Unmarshal([]byte(result), &chapters)
		if err != nil {
			log.Errorf("GetChapterByID: %v", err)
		}
	}
	if len(chapters) > 0 && role == "learner" {
		return chapters[0], nil
	} else {
		chapters = make([]*model.Chapter, 0)
	}

	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session
	qryStr := fmt.Sprintf(`SELECT * from coursez.chapter where id='%s' ALLOW FILTERING`, *chapterID)
	getChapters := func() (modules []coursez.Chapter, err error) {
		q := CassSession.Query(qryStr, nil)
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
	chaptersBytes, err := json.Marshal(chapters)
	if err != nil {
		log.Errorf("GetChapterByID: %v", err)
	} else {
		err = redis.SetRedisValue(ctx, key, string(chaptersBytes))
		if err != nil {
			log.Errorf("GetChapterByID: %v", err)
		}
		redis.SetTTL(ctx, key, 60)
	}
	return chapters[0], nil
}
