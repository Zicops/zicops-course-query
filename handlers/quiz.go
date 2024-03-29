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
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetTopicQuizes(ctx context.Context, topicID *string) ([]*model.Quiz, error) {
	topicQuizes := make([]*model.Quiz, 0)
	currentQuizes := make([]coursez.Quiz, 0)
	key := "GetTopicQuizes" + *topicID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &currentQuizes)
		if err == nil {
			log.Errorf("GetTopicQuizes from redis")
		}
	}
	if len(currentQuizes) <= 0 {
		session, err := global.CassPool.GetSession(ctx, "coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.quiz where topicid='%s' AND is_active=true  ALLOW FILTERING`, *topicID)
		getTopicQuiz := func() (quizes []coursez.Quiz, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return quizes, iter.Select(&quizes)
		}
		currentQuizes, err = getTopicQuiz()
		if err != nil {
			return nil, err
		}
	}
	topicQuizes = make([]*model.Quiz, len(currentQuizes))
	if len(currentQuizes) <= 0 {
		return topicQuizes, nil
	}
	var wg sync.WaitGroup
	for i, topQuiz := range currentQuizes {
		tt := topQuiz
		wg.Add(1)
		go func(mod coursez.Quiz, i int) {
			createdAt := strconv.FormatInt(mod.CreatedAt, 10)
			updatedAt := strconv.FormatInt(mod.UpdatedAt, 10)
			currentQ := &model.Quiz{
				ID:          &mod.ID,
				Name:        &mod.Name,
				Type:        &mod.Type,
				CreatedAt:   &createdAt,
				UpdatedAt:   &updatedAt,
				IsMandatory: &mod.IsMandatory,
				Sequence:    &mod.Sequence,
				TopicID:     &mod.TopicID,
				CourseID:    &mod.CourseID,
				QuestionID:  &mod.QuestionID,
				QbID:        &mod.QbId,
				Weightage:   &mod.Weightage,
				Category:    &mod.Category,
				StartTime:   &mod.StartTime,
			}

			topicQuizes[i] = currentQ
			wg.Done()
		}(tt, i)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentQuizes)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 60)
	}
	return topicQuizes, nil
}

func GetQuizFiles(ctx context.Context, quizID *string) ([]*model.QuizFile, error) {
	currentFiles := make([]coursez.QuizFile, 0)
	key := "GetQuizFiles" + *quizID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &currentFiles)
		if err != nil {
			log.Errorf("GetQuizFiles from redis")
		}
	}

	if len(currentFiles) <= 0 {
		session, err := global.CassPool.GetSession(ctx, "coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.quiz_file where quizid='%s' AND is_active=true ALLOW FILTERING`, *quizID)
		getQuizFiles := func() (files []coursez.QuizFile, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return files, iter.Select(&files)
		}
		currentFiles, err = getQuizFiles()
		if err != nil {
			return nil, err
		}
	}
	gproject := googleprojectlib.GetGoogleProjectID()
	quizFiles := make([]*model.QuizFile, len(currentFiles))
	if len(currentFiles) <= 0 {
		return quizFiles, nil
	}
	var wg sync.WaitGroup
	for i, file := range currentFiles {
		cc := file
		url := ""
		wg.Add(1)
		go func(mod coursez.QuizFile, i int) {
			if mod.BucketPath != "" {
				storageC := bucket.NewStorageHandler()
				err = storageC.InitializeStorageClient(ctx, gproject, mod.LspId)
				if err != nil {
					log.Errorf("Failed to initialize storage: %v", err.Error())
					return
				}
				url = storageC.GetSignedURLForObjectCache(ctx, mod.BucketPath)
			}
			currentFile := &model.QuizFile{
				Name:    &mod.Name,
				Type:    &mod.Type,
				QuizID:  &mod.QuizId,
				FileURL: &url,
			}

			quizFiles[i] = currentFile
			wg.Done()
		}(cc, i)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(currentFiles)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 60)
	}
	return quizFiles, nil
}

func GetMCQQuiz(ctx context.Context, quizID *string) ([]*model.QuizMcq, error) {
	quizMcqs := make([]*model.QuizMcq, 0)
	key := "GetMCQQuiz" + *quizID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &quizMcqs)
		if err == nil {
			return quizMcqs, nil
		}
	}

	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from coursez.quiz_mcq where quizid='%s' and is_active=true ALLOW FILTERING`, *quizID)
	getQuizMcq := func() (mcqs []coursez.QuizMcq, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return mcqs, iter.Select(&mcqs)
	}
	currentMCQs, err := getQuizMcq()
	if err != nil {
		return nil, err
	}
	quizMcqs = make([]*model.QuizMcq, len(currentMCQs))
	if len(currentMCQs) <= 0 {
		return quizMcqs, nil
	}
	var wg sync.WaitGroup
	for i, mcq := range currentMCQs {
		mm := mcq
		wg.Add(1)
		go func(mod coursez.QuizMcq, i int) {
			options := make([]*string, 0)
			for _, option := range mod.Options {
				options = append(options, &option)
			}
			currentMcq := &model.QuizMcq{
				QuizID:        &mod.QuizId,
				Explanation:   &mod.Explanation,
				Options:       options,
				Question:      &mod.Question,
				CorrectOption: &mod.CorrectOption,
			}
			quizMcqs[i] = currentMcq
			wg.Done()
		}(mm, i)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(quizMcqs)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 60)
	}
	return quizMcqs, nil
}

func GetQuizDes(ctx context.Context, quizID *string) ([]*model.QuizDescriptive, error) {
	quizDes := make([]*model.QuizDescriptive, 0)
	key := "GetQuizDes" + *quizID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &quizDes)
		if err == nil {
			return quizDes, nil
		}
	}
	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from coursez.quiz_descriptive where quizid='%s' AND is_active=true ALLOW FILTERING`, *quizID)
	getQuizDes := func() (desq []coursez.QuizDescriptive, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return desq, iter.Select(&desq)
	}
	currentDesQ, err := getQuizDes()
	if err != nil {
		return nil, err
	}
	quizDes = make([]*model.QuizDescriptive, len(currentDesQ))
	if len(currentDesQ) <= 0 {
		return quizDes, nil
	}
	var wg sync.WaitGroup
	for i, mcq := range currentDesQ {
		cd := mcq
		wg.Add(1)
		go func(mod coursez.QuizDescriptive, i int) {
			currentMcq := &model.QuizDescriptive{
				QuizID:        &mod.QuizId,
				Explanation:   &mod.Explanation,
				Question:      &mod.Question,
				CorrectAnswer: &mod.CorrectAnswer,
			}

			quizDes[i] = currentMcq
			wg.Done()
		}(cd, i)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(quizDes)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 60)
	}

	return quizDes, nil
}
