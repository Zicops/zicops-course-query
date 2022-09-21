package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetTopicQuizes(ctx context.Context, topicID *string) ([]*model.Quiz, error) {
	topicQuizes := make([]*model.Quiz, 0)
	currentQuizes := make([]coursez.Quiz, 0)
	key := "GetTopicQuizes" + *topicID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		err = json.Unmarshal([]byte(result), &currentQuizes)
		if err == nil {
			log.Errorf("GetTopicQuizes from redis")
		}
	}
	if len(currentQuizes) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.quiz where topicid='%s' ALLOW FILTERING`, *topicID)
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
	for _, topQuiz := range currentQuizes {
		mod := topQuiz
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

		topicQuizes = append(topicQuizes, currentQ)
	}
	redisBytes, err := json.Marshal(currentQuizes)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return topicQuizes, nil
}

func GetQuizFiles(ctx context.Context, quizID *string) ([]*model.QuizFile, error) {
	quizFiles := make([]*model.QuizFile, 0)
	currentFiles := make([]coursez.QuizFile, 0)
	key := "GetQuizFiles" + *quizID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		err = json.Unmarshal([]byte(result), &currentFiles)
		if err != nil {
			log.Errorf("GetQuizFiles from redis")
		}
	}

	if len(currentFiles) <= 0 {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from coursez.quiz_file where quizid='%s' ALLOW FILTERING`, *quizID)
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
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to initialize storage: %v", err.Error())
		return nil, err
	}
	for _, file := range currentFiles {
		mod := file
		url := ""
		if mod.BucketPath != "" {
			url = storageC.GetSignedURLForObject(mod.BucketPath)
		}
		currentFile := &model.QuizFile{
			Name:    &mod.Name,
			Type:    &mod.Type,
			QuizID:  &mod.QuizId,
			FileURL: &url,
		}

		quizFiles = append(quizFiles, currentFile)
	}
	redisBytes, err := json.Marshal(currentFiles)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return quizFiles, nil
}

func GetMCQQuiz(ctx context.Context, quizID *string) ([]*model.QuizMcq, error) {
	quizMcqs := make([]*model.QuizMcq, 0)
	key := "GetMCQQuiz" + *quizID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		err = json.Unmarshal([]byte(result), &quizMcqs)
		if err == nil {
			return quizMcqs, nil
		}
	}

	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from coursez.quiz_mcq where quizid='%s' ALLOW FILTERING`, *quizID)
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
	for _, mcq := range currentMCQs {
		mod := mcq
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

		quizMcqs = append(quizMcqs, currentMcq)
	}
	redisBytes, err := json.Marshal(quizMcqs)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return quizMcqs, nil
}

func GetQuizDes(ctx context.Context, quizID *string) ([]*model.QuizDescriptive, error) {
	quizDes := make([]*model.QuizDescriptive, 0)
	key := "GetQuizDes" + *quizID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		err = json.Unmarshal([]byte(result), &quizDes)
		if err == nil {
			return quizDes, nil
		}
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from coursez.quiz_descriptive where quizid='%s' ALLOW FILTERING`, *quizID)
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
	for _, mcq := range currentDesQ {
		mod := mcq
		currentMcq := &model.QuizDescriptive{
			QuizID:        &mod.QuizId,
			Explanation:   &mod.Explanation,
			Question:      &mod.Question,
			CorrectAnswer: &mod.CorrectAnswer,
		}

		quizDes = append(quizDes, currentMcq)
	}
	redisBytes, err := json.Marshal(quizDes)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}

	return quizDes, nil
}
