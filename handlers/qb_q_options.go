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
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetOptionsForQuestions(ctx context.Context, questionIds []*string) ([]*model.MapQuestionWithOption, error) {
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
	responseMap := make([]*model.MapQuestionWithOption, 0)
	for _, questionId := range questionIds {
		key := "GetOptionsForQuestions" + *questionId
		result, err := redis.GetRedisValue(key)
		banks := make([]qbankz.OptionsMain, 0)
		if err == nil && role != "admin" {
			err = json.Unmarshal([]byte(result), &banks)
			if err != nil {
				log.Errorf("Error in unmarshalling redis value: %v", err)
			}
		}
		currentMap := &model.MapQuestionWithOption{}
		currentMap.QuestionID = questionId

		if len(banks) <= 0 {
			qryStr := fmt.Sprintf(`SELECT * from qbankz.options_main where qm_id='%s' AND is_active=true ALLOW FILTERING`, *questionId)
			getBanks := func() (banks []qbankz.OptionsMain, err error) {
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
		allSections := make([]*model.QuestionOption, len(banks))
		if len(banks) <= 0 {
			currentMap.Options = allSections
			responseMap = append(responseMap, currentMap)
			continue
		}
		var wg sync.WaitGroup
		for i, bank := range banks {
			copiedQuestion := bank
			wg.Add(1)
			go func(copiedQuestion qbankz.OptionsMain, i int) {
				createdAt := strconv.FormatInt(copiedQuestion.CreatedAt, 10)
				updatedAt := strconv.FormatInt(copiedQuestion.UpdatedAt, 10)
				attUrl := ""

				if copiedQuestion.AttachmentBucket != "" {
					storageC := bucket.NewStorageHandler()
					gproject := googleprojectlib.GetGoogleProjectID()

					err := storageC.InitializeStorageClient(ctx, gproject, copiedQuestion.LspId)
					if err != nil {
						log.Errorf("Error in initializing storage client: %v", err)
					}
					attUrl = storageC.GetSignedURLForObject(copiedQuestion.AttachmentBucket)
				}
				currentQuestion := &model.QuestionOption{
					ID:             &copiedQuestion.ID,
					QmID:           &copiedQuestion.QmId,
					Description:    &copiedQuestion.Description,
					IsCorrect:      &copiedQuestion.IsCorrect,
					AttachmentType: &copiedQuestion.AttachmentType,
					CreatedBy:      &copiedQuestion.CreatedBy,
					CreatedAt:      &createdAt,
					UpdatedBy:      &copiedQuestion.UpdatedBy,
					UpdatedAt:      &updatedAt,
					IsActive:       &copiedQuestion.IsActive,
					Attachment:     &attUrl,
				}
				allSections[i] = currentQuestion
				wg.Done()
			}(copiedQuestion, i)
		}
		wg.Wait()
		currentMap.Options = allSections
		redisBytes, err := json.Marshal(banks)
		if err == nil {
			redis.SetTTL(key, 3600)
			redis.SetRedisValue(key, string(redisBytes))
		}
		responseMap = append(responseMap, currentMap)
	}

	return responseMap, nil
}
