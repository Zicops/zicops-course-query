package handlers

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/qbankz"
	"github.com/zicops/zicops-cass-pool/cassandra"
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
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from qbankz.question_bank_main where updated_at <= %d  ALLOW FILTERING`, *publishTime)
	getBanks := func(page []byte) (banks []qbankz.QuestionBankMain, nextPage []byte, err error) {
		q := CassSession.Query(qryStr, nil)
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
			Description: &copiedBank.Description,
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
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from qbankz.question_paper_main where updated_at <= %d  ALLOW FILTERING`, *publishTime)
	getBanks := func(page []byte) (banks []qbankz.QuestionPaperMain, nextPage []byte, err error) {
		q := CassSession.Query(qryStr, nil)
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
			Status:            &copiedBank.Status,
		}
		allBanks = append(allBanks, currentBank)
	}
	outputResponse.QuestionPapers = allBanks
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}

func GetLatestExams(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedExams, error) {
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
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam where updated_at <= %d  ALLOW FILTERING`, *publishTime)
	getExams := func(page []byte) (exams []qbankz.Exam, nextPage []byte, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return exams, iter.PageState(), iter.Select(&exams)
	}
	exams, newPage, err := getExams(newPage)
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
	var outputResponse model.PaginatedExams
	allExams := make([]*model.Exam, 0)
	for _, exam := range exams {
		copiedExam := exam
		createdAt := strconv.FormatInt(copiedExam.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedExam.UpdatedAt, 10)
		currentExam := &model.Exam{
			ID:           &copiedExam.ID,
			Name:         &copiedExam.Name,
			Description:  &copiedExam.Description,
			Code:         &copiedExam.Code,
			QpID:         &copiedExam.QPID,
			CreatedAt:    &createdAt,
			UpdatedAt:    &updatedAt,
			CreatedBy:    &copiedExam.CreatedBy,
			UpdatedBy:    &copiedExam.UpdatedBy,
			IsActive:     &copiedExam.IsActive,
			Type:         &copiedExam.Type,
			ScheduleType: &copiedExam.ScheduleType,
			Duration:     &copiedExam.Duration,
			Status:       &copiedExam.Status,
			Category:     &copiedExam.Category,
			SubCategory:  &copiedExam.SubCategory,
		}
		allExams = append(allExams, currentExam)
	}
	outputResponse.Exams = allExams
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}

func GetExamsMeta(ctx context.Context, examIds []*string) ([]*model.Exam, error) {
	responseMap := make([]*model.Exam, 0)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	for _, questionId := range examIds {
		qryStr := fmt.Sprintf(`SELECT * from qbankz.exam where id='%s'  ALLOW FILTERING`, *questionId)
		getPapers := func() (banks []qbankz.Exam, err error) {
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
			copiedExam := bank
			createdAt := strconv.FormatInt(copiedExam.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedExam.UpdatedAt, 10)
			currentExam := &model.Exam{
				ID:           &copiedExam.ID,
				Name:         &copiedExam.Name,
				Description:  &copiedExam.Description,
				Code:         &copiedExam.Code,
				QpID:         &copiedExam.QPID,
				CreatedAt:    &createdAt,
				UpdatedAt:    &updatedAt,
				CreatedBy:    &copiedExam.CreatedBy,
				UpdatedBy:    &copiedExam.UpdatedBy,
				IsActive:     &copiedExam.IsActive,
				Type:         &copiedExam.Type,
				ScheduleType: &copiedExam.ScheduleType,
				Duration:     &copiedExam.Duration,
				Status:       &copiedExam.Status,
				Category:     &copiedExam.Category,
				SubCategory:  &copiedExam.SubCategory,
			}
			responseMap = append(responseMap, currentExam)
		}
	}

	return responseMap, nil
}

func GetQBMeta(ctx context.Context, qbIds []*string) ([]*model.QuestionBank, error) {
	responseMap := make([]*model.QuestionBank, 0)
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	for _, qbId := range qbIds {
		qryStr := fmt.Sprintf(`SELECT * from qbankz.question_bank_main where id='%s'  ALLOW FILTERING`, *qbId)
		getBanks := func() (banks []qbankz.QuestionBankMain, err error) {
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
			copiedBank := bank
			createdAt := strconv.FormatInt(copiedBank.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedBank.UpdatedAt, 10)
			currentBank := &model.QuestionBank{
				ID:          &copiedBank.ID,
				Name:        &copiedBank.Name,
				Description: &copiedBank.Description,
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
			responseMap = append(responseMap, currentBank)
		}
	}
	return responseMap, nil
}
