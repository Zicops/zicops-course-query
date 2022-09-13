package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zicops/contracts/qbankz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
)

func GetQuestionBankSections(ctx context.Context, questionPaperID *string) ([]*model.QuestionPaperSection, error) {
	qryStr := fmt.Sprintf(`SELECT * from qbankz.section_main where qp_id = '%s'  ALLOW FILTERING`, *questionPaperID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	global.CassSession = session
	defer global.CassSession.Close()
	getBanks := func() (banks []qbankz.SectionMain, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSections := make([]*model.QuestionPaperSection, 0)
	for _, bank := range banks {
		copiedQuestion := bank
		createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
		currentQuestion := &model.QuestionPaperSection{
			ID:              &copiedQuestion.ID,
			Description:     &copiedQuestion.Description,
			Type:            &copiedQuestion.Type,
			CreatedBy:       &copiedQuestion.CreatedBy,
			CreatedAt:       &createdAt,
			UpdatedBy:       &copiedQuestion.UpdatedBy,
			UpdatedAt:       &updatedAt,
			Name:            &copiedQuestion.Name,
			DifficultyLevel: &copiedQuestion.DifficultyLevel,
			TotalQuestions:  &copiedQuestion.TotalQuestions,
			IsActive:        &copiedQuestion.IsActive,
			QpID:            &copiedQuestion.QPID,
		}
		allSections = append(allSections, currentQuestion)
	}
	return allSections, nil
}

func GetQPBankMappingByQPId(ctx context.Context, questionPaperID *string) ([]*model.SectionQBMapping, error) {
	qryStr := fmt.Sprintf(`SELECT * from qbankz.section_qb_mapping where qb_id = '%s'  ALLOW FILTERING`, *questionPaperID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	global.CassSession = session
	defer global.CassSession.Close()
	getBanks := func() (banks []qbankz.SectionQBMapping, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSectionsMap := make([]*model.SectionQBMapping, 0)
	for _, bank := range banks {
		copiedQuestion := bank
		createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
		currentQuestion := &model.SectionQBMapping{
			ID:              &copiedQuestion.ID,
			SectionID:       &copiedQuestion.SectionID,
			QbID:            &copiedQuestion.QBId,
			CreatedBy:       &copiedQuestion.CreatedBy,
			CreatedAt:       &createdAt,
			UpdatedBy:       &copiedQuestion.UpdatedBy,
			UpdatedAt:       &updatedAt,
			DifficultyLevel: &copiedQuestion.DifficultyLevel,
			TotalQuestions:  &copiedQuestion.TotalQuestions,
			IsActive:        &copiedQuestion.IsActive,
			QuestionMarks:   &copiedQuestion.QuestionMarks,
			QuestionType:    &copiedQuestion.QuestionType,
			RetrieveType:    &copiedQuestion.RetrievalType,
		}
		allSectionsMap = append(allSectionsMap, currentQuestion)
	}
	return allSectionsMap, nil
}

func GetQPBankMappingBySectionID(ctx context.Context, sectionID *string) ([]*model.SectionQBMapping, error) {
	qryStr := fmt.Sprintf(`SELECT * from qbankz.section_qb_mapping where section_id = '%s'  ALLOW FILTERING`, *sectionID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	global.CassSession = session
	defer global.CassSession.Close()
	getBanks := func() (banks []qbankz.SectionQBMapping, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSectionsMap := make([]*model.SectionQBMapping, 0)
	for _, bank := range banks {
		copiedQuestion := bank
		createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
		currentQuestion := &model.SectionQBMapping{
			ID:              &copiedQuestion.ID,
			SectionID:       &copiedQuestion.SectionID,
			QbID:            &copiedQuestion.QBId,
			CreatedBy:       &copiedQuestion.CreatedBy,
			CreatedAt:       &createdAt,
			UpdatedBy:       &copiedQuestion.UpdatedBy,
			UpdatedAt:       &updatedAt,
			DifficultyLevel: &copiedQuestion.DifficultyLevel,
			TotalQuestions:  &copiedQuestion.TotalQuestions,
			IsActive:        &copiedQuestion.IsActive,
			QuestionMarks:   &copiedQuestion.QuestionMarks,
			QuestionType:    &copiedQuestion.QuestionType,
			RetrieveType:    &copiedQuestion.RetrievalType,
		}
		allSectionsMap = append(allSectionsMap, currentQuestion)
	}
	return allSectionsMap, nil
}

func GetSectionFixedQuestions(ctx context.Context, sectionID *string) ([]*model.SectionFixedQuestions, error) {
	qryStr := fmt.Sprintf(`SELECT * from qbankz.section_fixed_questions where sqb_id = '%s'  ALLOW FILTERING`, *sectionID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	global.CassSession = session
	defer global.CassSession.Close()
	getBanks := func() (banks []qbankz.SectionFixedQuestions, err error) {
		q := global.CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	allSectionsMap := make([]*model.SectionFixedQuestions, 0)
	for _, bank := range banks {
		copiedQuestion := bank
		createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
		currentQuestion := &model.SectionFixedQuestions{
			ID:         &copiedQuestion.ID,
			SqbID:      &copiedQuestion.SQBId,
			QuestionID: &copiedQuestion.QuestionID,
			CreatedBy:  &copiedQuestion.CreatedBy,
			CreatedAt:  &createdAt,
			UpdatedBy:  &copiedQuestion.UpdatedBy,
			UpdatedAt:  &updatedAt,
			IsActive:   &copiedQuestion.IsActive,
		}
		allSectionsMap = append(allSectionsMap, currentQuestion)
	}
	return allSectionsMap, nil
}
