package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/zicops/contracts/qbankz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
)

func GetQuestionBankSections(ctx context.Context, questionPaperID *string) ([]*model.QuestionPaperSection, error) {
	allSections := make([]*model.QuestionPaperSection, 0)
	key := "GetQuestionBankSections" + *questionPaperID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
		err = json.Unmarshal([]byte(result), &allSections)
		if err == nil {
			return allSections, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.section_main where qp_id = '%s'  AND is_active=true ALLOW FILTERING`, *questionPaperID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.SectionMain, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
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
	redisBytes, err := json.Marshal(allSections)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return allSections, nil
}

func GetQPBankMappingByQPId(ctx context.Context, questionPaperID *string) ([]*model.SectionQBMapping, error) {
	key := "GetQPBankMappingByQPId" + *questionPaperID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
		allSectionsMap := make([]*model.SectionQBMapping, 0)
		err = json.Unmarshal([]byte(result), &allSectionsMap)
		if err == nil {
			return allSectionsMap, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.section_qb_mapping where qb_id = '%s'  AND is_active=true  ALLOW FILTERING`, *questionPaperID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.SectionQBMapping, err error) {
		q := CassSession.Query(qryStr, nil)
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
	redisBytes, err := json.Marshal(allSectionsMap)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return allSectionsMap, nil
}

func GetQPBankMappingBySectionID(ctx context.Context, sectionID *string) ([]*model.SectionQBMapping, error) {
	key := "GetQPBankMappingBySectionID" + *sectionID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
		allSectionsMap := make([]*model.SectionQBMapping, 0)
		err = json.Unmarshal([]byte(result), &allSectionsMap)
		if err == nil {
			return allSectionsMap, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.section_qb_mapping where section_id = '%s' AND is_active=true  ALLOW FILTERING`, *sectionID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.SectionQBMapping, err error) {
		q := CassSession.Query(qryStr, nil)
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
	redisBytes, err := json.Marshal(allSectionsMap)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return allSectionsMap, nil
}

func GetSectionFixedQuestions(ctx context.Context, sectionID *string) ([]*model.SectionFixedQuestions, error) {
	key := "GetSectionFixedQuestions" + *sectionID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(key)
	if err == nil && role != "admin" {
		allSectionsMap := make([]*model.SectionFixedQuestions, 0)
		err = json.Unmarshal([]byte(result), &allSectionsMap)
		if err == nil {
			return allSectionsMap, nil
		}
	}

	qryStr := fmt.Sprintf(`SELECT * from qbankz.section_fixed_questions where sqb_id = '%s' AND is_active=true  ALLOW FILTERING`, *sectionID)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	getBanks := func() (banks []qbankz.SectionFixedQuestions, err error) {
		q := CassSession.Query(qryStr, nil)
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
	redisBytes, err := json.Marshal(allSectionsMap)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return allSectionsMap, nil
}
