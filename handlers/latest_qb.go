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
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
)

func LatestQuestionBanks(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string, lspID *string) (*model.PaginatedQuestionBank, error) {
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
	key := "LatestQuestionBanks" + string(newPage)
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err != nil {
		log.Errorf("Error getting redis value: %v", err)
	}
	if result != "" && role == "learner" {
		var outputResponse model.PaginatedQuestionBank
		err = json.Unmarshal([]byte(result), &outputResponse)
		if err != nil {
			log.Errorf("Error unmarshalling redis value: %v", err)
		} else {
			return &outputResponse, nil
		}
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
	lspIdFromClaims := claims["lsp_id"].(string)
	whereClause := ""
	if searchText != nil && *searchText != "" {
		searchTextLower := strings.ToLower(*searchText)
		words := strings.Split(searchTextLower, " ")
		for _, word := range words {
			whereClause += " AND words CONTAINS '" + word + "'"
		}
	}
	//log.Println(lspIdFromClaims == *lspID)
	if lspID != nil && *lspID != "" {
		lspIdFromClaims = *lspID
	}
	qryStr := fmt.Sprintf(`SELECT * from qbankz.question_bank_main where created_at <= %d AND lsp_id='%s' AND is_active=true %s ALLOW FILTERING`, *publishTime, lspIdFromClaims, whereClause)
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
	redisBytes, err := json.Marshal(outputResponse)
	if err != nil {
		log.Errorf("Error marshalling redis value: %v", err)
	} else {
		redis.SetTTL(ctx, key, 60)
		err = redis.SetRedisValue(ctx, key, string(redisBytes))
		if err != nil {
			log.Errorf("Error setting redis value: %v", err)
		}
	}
	return &outputResponse, nil
}

func LatestQuestionPapers(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string) (*model.PaginatedQuestionPapers, error) {
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
	key := "LatestQuestionPapers" + string(newPage)
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err != nil {
		log.Errorf("Error getting redis value: %v", err)
	}
	if result != "" && role == "learner" {
		var outputResponse model.PaginatedQuestionPapers
		err = json.Unmarshal([]byte(result), &outputResponse)
		if err != nil {
			log.Errorf("Error unmarshalling redis value: %v", err)
		} else {
			return &outputResponse, nil
		}
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
	lspId := claims["lsp_id"].(string)
	whereClause := ""
	if searchText != nil && *searchText != "" {
		whereClause = whereClause + " AND  name='" + *searchText + "'"
	}
	qryStr := fmt.Sprintf(`SELECT * from qbankz.question_paper_main where created_at <= %d AND lsp_id = '%s' AND is_active=true %s ALLOW FILTERING`, *publishTime, lspId, whereClause)
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
	allBanks := make([]*model.QuestionPaper, len(banks))
	if len(banks) <= 0 {
		outputResponse.QuestionPapers = allBanks
		return &outputResponse, nil
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
			allBanks[i] = currentBank
			wg.Done()
		}(i, c)
	}
	wg.Wait()
	outputResponse.QuestionPapers = allBanks
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	redisBytes, err := json.Marshal(outputResponse)
	if err != nil {
		log.Errorf("Error marshalling redis value: %v", err)
	} else {
		redis.SetTTL(ctx, key, 60)
		err = redis.SetRedisValue(ctx, key, string(redisBytes))
		if err != nil {
			log.Errorf("Error setting redis value: %v", err)
		}
	}
	return &outputResponse, nil
}

func GetLatestExams(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string) (*model.PaginatedExams, error) {
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
	key := "GetLatestExams" + string(newPage)
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lspId := claims["lsp_id"].(string)
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	if err != nil {
		log.Errorf("Error getting redis value: %v", err)
	}
	if result != "" && role == "learner" {
		var outputResponse model.PaginatedExams
		err = json.Unmarshal([]byte(result), &outputResponse)
		if err != nil {
			log.Errorf("Error unmarshalling redis value: %v", err)
		} else {
			return &outputResponse, nil
		}
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
	whereClause := "WHERE "
	if searchText != nil && *searchText != "" {
		searchTextLower := strings.ToLower(*searchText)
		words := strings.Split(searchTextLower, " ")
		for i, word := range words {
			if i == 0 {
				whereClause = whereClause + " words CONTAINS" + " '" + word + "'"
			} else {
				whereClause = whereClause + " AND words CONTAINS" + " '" + word + "'"
			}
		}
		whereClause = whereClause + " AND "
	}
	whereClause = fmt.Sprintf(` %s is_active = true AND created_at <= %d AND lsp_id = '%s'`, whereClause, *publishTime, lspId)
	qryStr := fmt.Sprintf(`SELECT * from qbankz.exam %s ALLOW FILTERING`, whereClause)
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
	allExams := make([]*model.Exam, len(exams))
	if len(exams) <= 0 {
		outputResponse.Exams = allExams
		return &outputResponse, nil
	}
	var wg sync.WaitGroup
	for i, exam := range exams {
		c := exam
		wg.Add(1)
		go func(i int, copiedExam qbankz.Exam) {
			createdAt := strconv.FormatInt(copiedExam.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedExam.UpdatedAt, 10)
			questionIDs := make([]*string, 0)
			for _, questionID := range copiedExam.QuestionIDs {
				copiedQId := questionID
				questionIDs = append(questionIDs, &copiedQId)
			}
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
				QuestionIds:  questionIDs,
				TotalCount:   &copiedExam.TotalCount,
			}
			allExams[i] = currentExam
			wg.Done()
		}(i, c)
	}
	wg.Wait()
	outputResponse.Exams = allExams
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	redisBytes, err := json.Marshal(outputResponse)
	if err != nil {
		log.Errorf("Error marshalling redis value: %v", err)
	} else {
		redis.SetTTL(ctx, key, 60)
		err = redis.SetRedisValue(ctx, key, string(redisBytes))
		if err != nil {
			log.Errorf("Error setting redis value: %v", err)
		}
	}
	return &outputResponse, nil
}

func GetExamsMeta(ctx context.Context, examIds []*string) ([]*model.Exam, error) {
	session, err := cassandra.GetCassSession("qbankz")
	if err != nil {
		return nil, err
	}
	CassSession := session
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	responseMap := make([]*model.Exam, 0)
	for _, questionId := range examIds {
		result, _ := redis.GetRedisValue(ctx, "GetExamsMeta"+*questionId)
		if result != "" && role == "learner" {
			var outputResponse model.Exam
			err = json.Unmarshal([]byte(result), &outputResponse)
			if err == nil {
				responseMap = append(responseMap, &outputResponse)
				continue
			}
		}

		//lspId := claims["lsp_id"].(string)

		qryStr := fmt.Sprintf(`SELECT * from qbankz.exam where id='%s' AND is_active=true ALLOW FILTERING`, *questionId)
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
		if len(banks) <= 0 {
			continue
		}
		resExams := make([]*model.Exam, len(banks))
		var wg sync.WaitGroup
		for i, bank := range banks {
			c := bank
			wg.Add(1)
			go func(i int, copiedExam qbankz.Exam) {
				createdAt := strconv.FormatInt(copiedExam.CreatedAt, 10)
				updatedAt := strconv.FormatInt(copiedExam.UpdatedAt, 10)
				questionIDs := make([]*string, 0)
				for _, questionID := range copiedExam.QuestionIDs {
					copiedQId := questionID
					questionIDs = append(questionIDs, &copiedQId)
				}
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
					QuestionIds:  questionIDs,
					TotalCount:   &copiedExam.TotalCount,
				}
				resExams[i] = currentExam
				redisBytes, err := json.Marshal(currentExam)
				if err == nil {
					redis.SetTTL(ctx, "GetExamsMeta"+*questionId, 60)
					redis.SetRedisValue(ctx, "GetExamsMeta"+*questionId, string(redisBytes))
				}
				wg.Done()
			}(i, c)
		}
		wg.Wait()
		responseMap = append(responseMap, resExams...)
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
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	lspId := claims["lsp_id"].(string)

	for _, qbId := range qbIds {
		result, _ := redis.GetRedisValue(ctx, "GetQBMeta"+*qbId)
		if result != "" && role == "learner" {
			var outputResponse model.QuestionBank
			err = json.Unmarshal([]byte(result), &outputResponse)
			if err == nil {
				responseMap = append(responseMap, &outputResponse)
				continue
			}
		}

		qryStr := fmt.Sprintf(`SELECT * from qbankz.question_bank_main where id='%s' AND lsp_id='%s' AND is_active=true ALLOW FILTERING`, *qbId, lspId)
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
		resBanks := make([]*model.QuestionBank, len(banks))
		if len(banks) <= 0 {
			continue
		}
		var wg sync.WaitGroup
		for i, bank := range banks {
			c := bank
			wg.Add(1)
			go func(i int, copiedBank qbankz.QuestionBankMain) {
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
				resBanks[i] = currentBank
				redisBytes, err := json.Marshal(currentBank)
				if err == nil {
					redis.SetTTL(ctx, "GetQBMeta"+*qbId, 60)
					redis.SetRedisValue(ctx, "GetQBMeta"+*qbId, string(redisBytes))
				}
				wg.Done()
			}(i, c)
		}
		wg.Wait()
		responseMap = append(responseMap, resBanks...)
	}
	return responseMap, nil
}
