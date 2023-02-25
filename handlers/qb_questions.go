package handlers

import (
	"context"
	"encoding/base64"
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
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetQuestionBankQuestions(ctx context.Context, questionBankID *string, filters *model.QBFilters) ([]*model.QuestionBankQuestion, error) {
	gproject := googleprojectlib.GetGoogleProjectID()

	key := "GetQuestionBankQuestions" + *questionBankID + fmt.Sprintf("%v", filters)
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	result, err := redis.GetRedisValue(ctx, key)
	banks := make([]qbankz.QuestionMain, 0)
	if err == nil && role == "learner" {
		err = json.Unmarshal([]byte(result), &banks)
		if err != nil {
			log.Errorf("Failed to unmarshal redis value: %v", err.Error())
		}
	}
	if len(banks) <= 0 {
		whereClause := getWhereClause(filters, *questionBankID)
		session, err := cassandra.GetCassSession("qbankz")
		if err != nil {
			return nil, err
		}
		CassSession := session

		qryStr := fmt.Sprintf(`SELECT * from qbankz.question_main where %s  ALLOW FILTERING`, whereClause)
		getBanks := func() (banks []qbankz.QuestionMain, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return banks, iter.Select(&banks)
		}
		banks, err = getBanks()
		if err != nil {
			return nil, err
		}
	}
	filteredBanks := make([]qbankz.QuestionMain, 0)
	excludedIds := []string{}
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
	allQuestions := make([]*model.QuestionBankQuestion, len(shuffledBanks))
	if len(shuffledBanks) <= 0 {
		return allQuestions, nil
	}
	var wg sync.WaitGroup
	for i, b := range shuffledBanks {
		bb := b
		wg.Add(1)
		go func(i int, bank qbankz.QuestionMain) {
			copiedQuestion := bank
			createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
			bucketQ := copiedQuestion.AttachmentBucket
			attUrl := ""
			if bucketQ != "" {
				key := base64.StdEncoding.EncodeToString([]byte(bucketQ))
				redisValue, err := redis.GetRedisValue(ctx, key)
				if err == nil && redisValue != ""{
					attUrl = redisValue
				} else {
					storageC := bucket.NewStorageHandler()
					err = storageC.InitializeStorageClient(ctx, gproject, copiedQuestion.LspId)
					if err != nil {
						log.Errorf("Failed to initialize storage client: %v", err.Error())
					}
					attUrl = storageC.GetSignedURLForObject(bucketQ)
					redis.SetRedisValue(ctx, key, attUrl)
					redis.SetTTL(ctx, key, 3000)
				}
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
			allQuestions[i] = currentQuestion
			wg.Done()
		}(i, bb)
	}
	wg.Wait()
	redisBytes, err := json.Marshal(banks)
	if err == nil {
		redis.SetTTL(ctx, key, 60)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return allQuestions, nil
}

func GetQuestionsByID(ctx context.Context, questionIds []*string) ([]*model.QuestionBankQuestion, error) {
	gproject := googleprojectlib.GetGoogleProjectID()

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
	allQuestions := make([]*model.QuestionBankQuestion, 0)

	for _, id := range questionIds {
		key := "GetQuestionsByID" + *id
		result, err := redis.GetRedisValue(ctx, key)
		banks := make([]qbankz.QuestionMain, 0)
		if err == nil && role == "learner" {
			json.Unmarshal([]byte(result), &banks)
		}
		if len(banks) <= 0 {
			qryStr := fmt.Sprintf(`SELECT * from qbankz.question_main where id = '%s' AND is_active=true ALLOW FILTERING`, *id)
			getBanks := func() (banks []qbankz.QuestionMain, err error) {
				q := CassSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return banks, iter.Select(&banks)
			}
			banks, err = getBanks()
			if err != nil {
				return nil, err
			}
		}
		if len(banks) <= 0 {
			continue
		}
		var wg sync.WaitGroup
		for i, bank := range banks {
			collectQs := make([]*model.QuestionBankQuestion, len(banks))
			c := bank
			wg.Add(1)
			go func(i int, copiedQuestion qbankz.QuestionMain) {
				createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
				updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
				bucketQ := copiedQuestion.AttachmentBucket
				attUrl := ""
				if bucketQ != "" {
					storageC := bucket.NewStorageHandler()
					err := storageC.InitializeStorageClient(ctx, gproject, copiedQuestion.LspId)
					if err != nil {
						log.Errorf("Failed to get questions: %v", err.Error())
						return
					}
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
				collectQs[i] = currentQuestion
				wg.Done()
			}(i, c)
			wg.Wait()
			allQuestions = append(allQuestions, collectQs...)
		}
		wg.Wait()
		redisBytes, err := json.Marshal(banks)
		if err == nil {
			redis.SetTTL(ctx, key, 60)
			redis.SetRedisValue(ctx, key, string(redisBytes))
		}
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
	whereClause = whereClause + fmt.Sprintf(" AND is_active = true")
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
