package handlers

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/qbankz"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
)

func LatestQuestionBanks(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedQuestionBank, error) {
	var newPage []byte
	//var pageDirection string
	var pageSizeInt int
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid page cursor: %v", err)
		}
		newPage = page
	}
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string

	qryStr := fmt.Sprintf(`SELECT * from qbankz.question_bank_main where updated_at <= %d  ALLOW FILTERING`, *publishTime)
	getBanks := func(page []byte) (banks []qbankz.QuestionBankMain, nextPage []byte, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return banks, iter.PageState(), iter.Select(&banks)
	}
	banks, newPage, err := getBanks(newPage)
	if err != nil {
		return nil, err
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Courses: %v", string(newCursor))

	}
	var outputResponse model.PaginatedQuestionBank
	allBanks := make([]*model.QuestionBank, 0)
	for _, bank := range banks {
		copiedBank := bank
		createdAt := strconv.FormatInt(copiedBank.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedBank.UpdatedAt, 10)
		currentBank := &model.QuestionBank{
			ID:          &copiedBank.ID,
			Name:        &copiedBank.Name,
			Category:    &copiedBank.Category,
			SubCategory: &copiedBank.SubCategory,
			Owner:       &copiedBank.Owner,
			IsActive:    &copiedBank.IsActive,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			CreatedBy:   &copiedBank.CreatedBy,
			UpdatedBy:   &copiedBank.UpdatedBy,
			IsDefault:   &copiedBank.IsDefault,
		}
		allBanks = append(allBanks, currentBank)
	}
	outputResponse.QuestionBanks = allBanks
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}

func LatestQuestionPapers(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedQuestionPapers, error) {
	var newPage []byte
	//var pageDirection string
	var pageSizeInt int
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid page cursor: %v", err)
		}
		newPage = page
	}
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string

	qryStr := fmt.Sprintf(`SELECT * from qbankz.question_paper_main where updated_at <= %d  ALLOW FILTERING`, *publishTime)
	getBanks := func(page []byte) (banks []qbankz.QuestionPaperMain, nextPage []byte, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return banks, iter.PageState(), iter.Select(&banks)
	}
	banks, newPage, err := getBanks(newPage)
	if err != nil {
		return nil, err
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Courses: %v", string(newCursor))

	}
	var outputResponse model.PaginatedQuestionPapers
	allBanks := make([]*model.QuestionPaper, 0)
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
		}
		allBanks = append(allBanks, currentBank)
	}
	outputResponse.QuestionPapers = allBanks
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}
