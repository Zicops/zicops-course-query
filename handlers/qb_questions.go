package handlers

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/qbankz"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetQuestionBankQuestions(ctx context.Context, questionBankID *string, filters *model.QBFilters) ([]*model.QuestionBankQuestion, error) {
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err := storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to get questions: %v", err.Error())
		return nil, err
	}

	whereClause := getWhereClause(filters, *questionBankID)

	qryStr := fmt.Sprintf(`SELECT * from qbankz.question_main where %s  ALLOW FILTERING`, whereClause)
	getBanks := func() (banks []qbankz.QuestionMain, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	banks, err := getBanks()
	if err != nil {
		return nil, err
	}
	filteredBanks := make([]qbankz.QuestionMain, 0)
	excludedIds := []string{}
	allQuestions := make([]*model.QuestionBankQuestion, 0)
	if filters != nil && filters.ExcludedQuestionIds != nil {
		currentExclusion := filters.ExcludedQuestionIds
		for _, id := range currentExclusion {
			copiedId := *id
			excludedIds = append(excludedIds, copiedId)
		}
		// filter out the questions that are in the excluded list
		for _, bank := range banks {
			if !contains(excludedIds, bank.ID) {
				filteredBanks = append(filteredBanks, bank)
			}
		}
	} else {
		filteredBanks = banks
	}
	shuffledBanks := shuffle(filteredBanks)
	if filters != nil && filters.TotalQuestions != nil {
		totalQuestions := *filters.TotalQuestions
		if totalQuestions > len(shuffledBanks) {
			totalQuestions = len(shuffledBanks)
		}
		shuffledBanks = shuffledBanks[:totalQuestions]
	}
	for _, bank := range shuffledBanks {
		copiedQuestion := bank
		createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
		bucketQ := copiedQuestion.AttachmentBucket
		attUrl := ""
		if bucketQ != "" {
			attUrl = storageC.GetSignedURLForObject(bucketQ)
		}
		currentQuestion := &model.QuestionBankQuestion{
			ID:             &copiedQuestion.ID,
			Name:           &copiedQuestion.Name,
			Description:    &copiedQuestion.Description,
			Type:           &copiedQuestion.Type,
			AttachmentType: &copiedQuestion.AttachmentType,
			Attachment:     &attUrl,
			Hint:           &copiedQuestion.Hint,
			Difficulty:     &copiedQuestion.Difficulty,
			QbmID:          &copiedQuestion.QbmId,
			Status:         &copiedQuestion.Status,
			CreatedBy:      &copiedQuestion.CreatedBy,
			CreatedAt:      &createdAt,
			UpdatedBy:      &copiedQuestion.UpdatedBy,
			UpdatedAt:      &updatedAt,
		}
		allQuestions = append(allQuestions, currentQuestion)
	}
	return allQuestions, nil
}

func getWhereClause(filters *model.QBFilters, qb_id string) string {
	whereClause := fmt.Sprintf("qbm_id = '%s'", qb_id)
	if filters != nil {
		if filters.DifficultyStart != nil {
			whereClause = fmt.Sprintf("%s AND difficulty_score >= %d", whereClause, *filters.DifficultyStart)
		}
		if filters.DifficultyEnd != nil {
			whereClause = fmt.Sprintf("%s AND difficulty_score <= %d", whereClause, *filters.DifficultyEnd)
		}
	}
	return whereClause
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func shuffle(s []qbankz.QuestionMain) []qbankz.QuestionMain {
	for i := len(s) - 1; i > 0; i-- {
		j := global.Rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
	return s
}
