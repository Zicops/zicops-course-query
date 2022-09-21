package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/zicops/contracts/qbankz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
)

func GetExamsByQPId(ctx context.Context, questionPaperID *string) ([]*model.Exam, error) {
	key := "GetExamsByQPId" + *questionPaperID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		output := make([]*model.Exam, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}
	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam where qp_id = '%s'  ALLOW FILTERING`, *questionPaperID)
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
	allSections := make([]*model.Exam, 0)
	for _, bank := range banks {
		copiedQuestion := bank
		createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
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
		}
		allSections = append(allSections, currentQuestion)
	}
	output, err := json.Marshal(allSections)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(output))
	}
	return allSections, nil
}

func GetExamSchedule(ctx context.Context, examID *string) ([]*model.ExamSchedule, error) {
	key := "GetExamSchedule" + *examID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		output := make([]*model.ExamSchedule, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_schedule where exam_id = '%s'  ALLOW FILTERING`, *examID)
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
	allSections := make([]*model.ExamSchedule, 0)
	for _, bank := range banks {
		copiedQuestion := bank
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
		allSections = append(allSections, currentQuestion)
	}
	output, err := json.Marshal(allSections)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(output))
	}
	return allSections, nil
}

func GetExamInstruction(ctx context.Context, examID *string) ([]*model.ExamInstruction, error) {
	key := "GetExamInstruction" + *examID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		output := make([]*model.ExamInstruction, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}
	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_instructions where exam_id = '%s'  ALLOW FILTERING`, *examID)
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
	allSections := make([]*model.ExamInstruction, 0)
	for _, bank := range banks {
		copiedQuestion := bank
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
		allSections = append(allSections, currentQuestion)
	}
	output, err := json.Marshal(allSections)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(output))
	}
	return allSections, nil
}

func GetExamCohort(ctx context.Context, examID *string) ([]*model.ExamCohort, error) {
	key := "GetExamCohort" + *examID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		output := make([]*model.ExamCohort, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_cohort where exam_id = '%s'  ALLOW FILTERING`, *examID)
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
	allSections := make([]*model.ExamCohort, 0)
	for _, bank := range banks {
		copiedQuestion := bank
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
		allSections = append(allSections, currentQuestion)
	}
	output, err := json.Marshal(allSections)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(output))
	}
	return allSections, nil
}

func GetExamConfiguration(ctx context.Context, examID *string) ([]*model.ExamConfiguration, error) {
	key := "GetExamConfiguration" + *examID
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := redis.GetRedisValue(key)
	if err == nil {
		output := make([]*model.ExamConfiguration, 0)
		err = json.Unmarshal([]byte(result), &output)
		if err == nil {
			return output, nil
		}
	}
	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_config where exam_id = '%s'  ALLOW FILTERING`, *examID)
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
	allSections := make([]*model.ExamConfiguration, 0)
	for _, bank := range banks {
		copiedQuestion := bank
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
		allSections = append(allSections, currentQuestion)
	}
	output, err := json.Marshal(allSections)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(output))
	}
	return allSections, nil
}

func GetQPMeta(ctx context.Context, questionPapersIds []*string) ([]*model.QuestionPaper, error) {
	responseMap := make([]*model.QuestionPaper, 0)
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	for _, questionId := range questionPapersIds {
		key := "GetQPMeta" + *questionId
		result, err := redis.GetRedisValue(key)
		if err == nil {
			output := &model.QuestionPaper{}
			err = json.Unmarshal([]byte(result), output)
			if err == nil {
				responseMap = append(responseMap, output)
				continue
			}
		}
		currentMap := &model.QuestionPaper{}
		currentMap.ID = questionId
		qryStr := fmt.Sprintf(`SELECT * from qbankz.question_paper_main where id='%s'  ALLOW FILTERING`, *questionId)
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
		for _, bank := range banks {
			copiedBank := bank
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
			responseMap = append(responseMap, currentBank)
			output, err := json.Marshal(currentBank)
			if err == nil {
				redis.SetTTL(key, 3600)
				redis.SetRedisValue(key, string(output))
			}
		}
	}

	return responseMap, nil
}
