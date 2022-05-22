package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zicops/contracts/qbankz"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
)

func GetExamsByQPId(ctx context.Context, questionPaperID *string) ([]*model.Exam, error) {
	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam where qp_id = %s  ALLOW FILTERING`, *questionPaperID)
	getBanks := func() (banks []qbankz.Exam, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
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
		}
		allSections = append(allSections, currentQuestion)
	}
	return allSections, nil
}

func GetExamSchedule(ctx context.Context, examID *string) (*model.ExamSchedule, error) {
	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam_schedule where exam_id = %s  ALLOW FILTERING`, *examID)
	getBanks := func() (banks []qbankz.ExamSchedule, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
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
	return allSections[0], nil
}
