package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/qbankz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
)

func GetExamsByQPId(ctx context.Context, questionPaperID *string) ([]*model.Exam, error) {
	key := "GetExamsByQPId" + *questionPaperID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		output := make([]*model.Exam, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam where qp_id = '%s'  AND is_active=true  ALLOW FILTERING`, *questionPaperID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.Exam, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSections := make([]*model.Exam, len(banks))
	if len(banks) <= 0 {
		return allSections, nil
	}
	var wg sync.WaitGroup
	for i, bank := range banks {
		cqb := bank
		wg.Add(1)
		go func(i int, copiedQuestion qbankz.Exam) {
			createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
			questionIDs := make([]*string, 0)
			for _, questionID := range copiedQuestion.QuestionIDs {
				copiedQId := questionID
				questionIDs = append(questionIDs, &copiedQId)
			}
			currentQuestion := &model.Exam{
				ID:           &copiedQuestion.ID,
				Description:  &copiedQuestion.Description,
				Type:         &copiedQuestion.Type,
				CreatedBy:    &copiedQuestion.CreatedBy,
				CreatedAt:    &createdAt,
				UpdatedBy:    &copiedQuestion.UpdatedBy,
				UpdatedAt:    &updatedAt,
				Name:         &copiedQuestion.Name,
				Code:         &copiedQuestion.Code,
				Category:     &copiedQuestion.Category,
				ScheduleType: &copiedQuestion.ScheduleType,
				SubCategory:  &copiedQuestion.SubCategory,
				Duration:     &copiedQuestion.Duration,
				Status:       &copiedQuestion.Status,
				IsActive:     &copiedQuestion.IsActive,
				QpID:         &copiedQuestion.QPID,
				QuestionIds:  questionIDs,
				TotalCount:   &copiedQuestion.TotalCount,
			}
			allSections[i] = currentQuestion
			wg.Done()
		}(i, cqb)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(allSections)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 60)
	}
	return allSections, nil
}

func GetExamSchedule(ctx context.Context, examID *string) ([]*model.ExamSchedule, error) {
	key := "GetExamSchedule" + *examID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		output := make([]*model.ExamSchedule, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_schedule where exam_id = '%s'  AND is_active=true  ALLOW FILTERING`, *examID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.ExamSchedule, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSections := make([]*model.ExamSchedule, len(banks))
	if len(banks) <= 0 {
		return allSections, nil
	}
	var wg sync.WaitGroup
	for i, bank := range banks {
		cqb := bank
		wg.Add(1)
		go func(i int, copiedQuestion qbankz.ExamSchedule) {
			createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
			bufferTimeInt := strconv.Itoa(copiedQuestion.BufferTime)
			startInt := strconv.FormatInt(copiedQuestion.Start, 10)
			endInt := strconv.FormatInt(copiedQuestion.End, 10)
			currentQuestion := &model.ExamSchedule{
				ID:         &copiedQuestion.ID,
				CreatedBy:  &copiedQuestion.CreatedBy,
				CreatedAt:  &createdAt,
				UpdatedBy:  &copiedQuestion.UpdatedBy,
				UpdatedAt:  &updatedAt,
				ExamID:     &copiedQuestion.ExamID,
				BufferTime: &bufferTimeInt,
				Start:      &startInt,
				End:        &endInt,
				IsActive:   &copiedQuestion.IsActive,
			}
			allSections[i] = currentQuestion
			wg.Done()
		}(i, cqb)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(allSections)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 60)
	}
	return allSections, nil
}

func GetExamScheduleByExamID(ctx context.Context, examIds []*string) ([]*model.ExamSchedule, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))

	key := "GetExamScheduleByExamID"
	for _, vv := range examIds {
		v := *vv
		key = key + v
	}
	output := make([]qbankz.ExamSchedule, 0)
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" && result != "" {
		err = json.Unmarshal([]byte(result), &output)
		if err != nil {
			return nil, err
		}
	}
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	//if unable to get data from cache, call the database
	if len(output) == 0 {
		for _, vv := range examIds {
			v := *vv
			qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_schedule where exam_id = '%s'  AND is_active=true  ALLOW FILTERING`, v)
			getSchedule := func() (schedules []qbankz.ExamSchedule, err error) {
				q := CassSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return schedules, iter.Select(&schedules)
			}

			examSchedules, err := getSchedule()
			if err != nil {
				log.Printf("Got error while getting schedule: %v", err)
				return nil, err
			}

			output = append(output, examSchedules...)
		}
	}
	res := make([]*model.ExamSchedule, len(output))
	var wg sync.WaitGroup

	for kk, vvv := range output {
		vv := vvv
		wg.Add(1)
		go func(k int, v qbankz.ExamSchedule) {
			createdAt := strconv.FormatInt(v.CreatedAt, 10)
			updatedAt := strconv.FormatInt(v.UpdatedAt, 10)
			bufferTimeInt := strconv.Itoa(v.BufferTime)
			startInt := strconv.FormatInt(v.Start, 10)
			endInt := strconv.FormatInt(v.End, 10)
			currentQuestion := &model.ExamSchedule{
				ID:         &v.ID,
				CreatedBy:  &v.CreatedBy,
				CreatedAt:  &createdAt,
				UpdatedBy:  &v.UpdatedBy,
				UpdatedAt:  &updatedAt,
				ExamID:     &v.ExamID,
				BufferTime: &bufferTimeInt,
				Start:      &startInt,
				End:        &endInt,
				IsActive:   &v.IsActive,
			}
			res[k] = currentQuestion
			wg.Done()
		}(kk, vv)
	}
	wg.Wait()

	tmp, err := json.Marshal(output)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(tmp))
		redis.SetTTL(ctx, key, 60)
	}

	return res, nil
}

func GetExamInstruction(ctx context.Context, examID *string) ([]*model.ExamInstruction, error) {
	key := "GetExamInstruction" + *examID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		output := make([]*model.ExamInstruction, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_instructions where exam_id = '%s' AND is_active=true   ALLOW FILTERING`, *examID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.ExamInstructions, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSections := make([]*model.ExamInstruction, len(banks))
	if len(banks) <= 0 {
		return allSections, nil
	}
	var wg sync.WaitGroup
	for i, bank := range banks {
		c := bank
		wg.Add(1)
		go func(i int, copiedQuestion qbankz.ExamInstructions) {
			createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
			attempts := strconv.Itoa(copiedQuestion.NoAttempts)
			currentQuestion := &model.ExamInstruction{
				ID:              &copiedQuestion.ID,
				Instructions:    &copiedQuestion.Instructions,
				CreatedBy:       &copiedQuestion.CreatedBy,
				CreatedAt:       &createdAt,
				UpdatedBy:       &copiedQuestion.UpdatedBy,
				UpdatedAt:       &updatedAt,
				ExamID:          &copiedQuestion.ExamID,
				IsActive:        &copiedQuestion.IsActive,
				PassingCriteria: &copiedQuestion.PassingCriteria,
				AccessType:      &copiedQuestion.AccessType,
				NoAttempts:      &attempts,
			}
			allSections[i] = currentQuestion
			wg.Done()
		}(i, c)
	}
	wg.Wait()
	output, err := json.Marshal(allSections)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(output))
		redis.SetTTL(ctx, key, 60)
	}
	return allSections, nil
}

func GetExamInstructionByExamID(ctx context.Context, examIds []*string) ([]*model.ExamInstruction, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	CurrentInstruction := make([]qbankz.ExamInstructions, 0)
	role := strings.ToLower(claims["role"].(string))
	key := "GetExamInstructionByExamID"
	for _, vv := range examIds {
		v := *vv
		key = key + v
	}
	result, err := redis.GetRedisValue(ctx, key)
	if (err == nil && role == "learner") && result != "" {

		err = json.Unmarshal([]byte(result), &CurrentInstruction)
		if err == nil {
			log.Errorf("Failed to unmarshal instructions: %v", err.Error())
			return nil, err
		}
	}

	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	if len(CurrentInstruction) == 0 {
		for _, vv := range examIds {
			v := *vv
			queryStr := fmt.Sprintf(`SELECT * from qbankz.exam_instructions where exam_id = '%s' AND is_active=true   ALLOW FILTERING`, v)
			getBanks := func() (banks []qbankz.ExamInstructions, err error) {
				q := CassSession.Query(queryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return banks, iter.Select(&banks)
			}
			banks, err := getBanks()
			if err != nil {
				return nil, err
			}
			if len(banks) == 0 {
				continue
			}
			CurrentInstruction = append(CurrentInstruction, banks...)
		}
	}

	res := make([]*model.ExamInstruction, len(CurrentInstruction))
	var wg sync.WaitGroup
	for kk, vvv := range CurrentInstruction {
		vv := vvv
		wg.Add(1)

		go func(k int, v qbankz.ExamInstructions) {
			createdAt := strconv.FormatInt(v.CreatedAt, 10)
			updatedAt := strconv.FormatInt(v.UpdatedAt, 10)
			attempts := strconv.Itoa(v.NoAttempts)
			tmp := &model.ExamInstruction{
				ID:              &v.ID,
				ExamID:          &v.ExamID,
				Instructions:    &v.Instructions,
				PassingCriteria: &v.PassingCriteria,
				NoAttempts:      &attempts,
				AccessType:      &v.AccessType,
				CreatedAt:       &createdAt,
				UpdatedAt:       &updatedAt,
				CreatedBy:       &v.CreatedBy,
				UpdatedBy:       &v.UpdatedBy,
				IsActive:        &v.IsActive,
			}

			res[k] = tmp
			wg.Done()
		}(kk, vv)
	}
	wg.Wait()
	output, err := json.Marshal(CurrentInstruction)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(output))
		redis.SetTTL(ctx, key, 60)
	}
	return res, nil

}

func GetExamCohort(ctx context.Context, examID *string) ([]*model.ExamCohort, error) {
	key := "GetExamCohort" + *examID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		output := make([]*model.ExamCohort, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_cohort where exam_id = '%s' AND is_active=true  ALLOW FILTERING`, *examID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.ExamCohort, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSections := make([]*model.ExamCohort, len(banks))
	if len(banks) <= 0 {
		return allSections, nil
	}
	var wg sync.WaitGroup
	for i, bank := range banks {
		c := bank
		wg.Add(1)
		go func(i int, copiedQuestion qbankz.ExamCohort) {
			createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
			currentQuestion := &model.ExamCohort{
				ID:        &copiedQuestion.ID,
				CreatedBy: &copiedQuestion.CreatedBy,
				CreatedAt: &createdAt,
				UpdatedBy: &copiedQuestion.UpdatedBy,
				UpdatedAt: &updatedAt,
				ExamID:    &copiedQuestion.ExamID,
				IsActive:  &copiedQuestion.IsActive,
				CohortID:  &copiedQuestion.CohortID,
			}
			allSections[i] = currentQuestion
			wg.Done()
		}(i, c)
	}
	wg.Wait()
	output, err := json.Marshal(allSections)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(output))
		redis.SetTTL(ctx, key, 60)
	}
	return allSections, nil
}

func GetExamConfiguration(ctx context.Context, examID *string) ([]*model.ExamConfiguration, error) {
	key := "GetExamConfiguration" + *examID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && role == "learner" {
		output := make([]*model.ExamConfiguration, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_config where exam_id = '%s' AND is_active=true  ALLOW FILTERING`, *examID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.ExamConfig, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSections := make([]*model.ExamConfiguration, len(banks))
	if len(banks) <= 0 {
		return allSections, nil
	}
	var wg sync.WaitGroup
	for i, bank := range banks {
		c := bank
		wg.Add(1)
		go func(i int, copiedQuestion qbankz.ExamConfig) {
			createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
			currentQuestion := &model.ExamConfiguration{
				ID:           &copiedQuestion.ID,
				CreatedBy:    &copiedQuestion.CreatedBy,
				CreatedAt:    &createdAt,
				UpdatedBy:    &copiedQuestion.UpdatedBy,
				UpdatedAt:    &updatedAt,
				ExamID:       &copiedQuestion.ExamID,
				IsActive:     &copiedQuestion.IsActive,
				Shuffle:      &copiedQuestion.Shuffle,
				DisplayHints: &copiedQuestion.DisplayHints,
				ShowAnswer:   &copiedQuestion.ShowAnswer,
				ShowResult:   &copiedQuestion.ShowResult,
			}
			allSections[i] = currentQuestion
			wg.Done()
		}(i, c)
	}
	wg.Wait()
	output, err := json.Marshal(allSections)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(output))
		redis.SetTTL(ctx, key, 60)
	}
	return allSections, nil
}

func GetQPMeta(ctx context.Context, questionPapersIds []*string) ([]*model.QuestionPaper, error) {
	responseMap := make([]*model.QuestionPaper, 0)
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	for _, questionId := range questionPapersIds {
		key := "GetQPMeta" + *questionId
		result, err := redis.GetRedisValue(ctx, key)
		if err == nil && role == "learner" {
			output := &model.QuestionPaper{}
			err = json.Unmarshal([]byte(result), output)
			if err == nil {
				responseMap = append(responseMap, output)
				continue
			}
		}

		qryStr := fmt.Sprintf(`SELECT * from qbankz.question_paper_main where id='%s'  AND is_active=true  ALLOW FILTERING`, *questionId)
		getPapers := func() (banks []qbankz.QuestionPaperMain, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return banks, iter.Select(&banks)
		}
		banks, err := getPapers()
		if err != nil {
			return nil, err
		}
		resCurrent := make([]*model.QuestionPaper, len(banks))
		if len(banks) <= 0 {
			continue
		}
		var wg sync.WaitGroup
		for i, bank := range banks {
			c := bank
			wg.Add(1)
			go func(i int, copiedBank qbankz.QuestionPaperMain) {
				createdAt := strconv.FormatInt(copiedBank.CreatedAt, 10)
				updatedAt := strconv.FormatInt(copiedBank.UpdatedAt, 10)
				currentBank := &model.QuestionPaper{
					ID:                &copiedBank.ID,
					Name:              &copiedBank.Name,
					Category:          &copiedBank.Category,
					SubCategory:       &copiedBank.SubCategory,
					SuggestedDuration: &copiedBank.SuggestedDuration,
					SectionWise:       &copiedBank.SectionWise,
					DifficultyLevel:   &copiedBank.DifficultyLevel,
					Description:       &copiedBank.Description,
					IsActive:          &copiedBank.IsActive,
					CreatedAt:         &createdAt,
					UpdatedAt:         &updatedAt,
					CreatedBy:         &copiedBank.CreatedBy,
					UpdatedBy:         &copiedBank.UpdatedBy,
					Status:            &copiedBank.Status,
				}
				resCurrent[i] = currentBank
				output, err := json.Marshal(currentBank)
				if err == nil {
					redis.SetRedisValue(ctx, key, string(output))
					redis.SetTTL(ctx, key, 60)
				}
				wg.Done()
			}(i, c)
		}
		wg.Wait()
		responseMap = append(responseMap, resCurrent...)
	}

	return responseMap, nil
}
