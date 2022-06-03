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

func GetQuestionBankQuestions(ctx context.Context, questionBankID *string) ([]*model.QuestionBankQuestion, error) {
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err := storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to get questions: %v", err.Error())
		return nil, err
	}
	qryStr := fmt.Sprintf(`SELECT * from qbankz.question_main where qbm_id = '%s'  ALLOW FILTERING`, *questionBankID)
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
	allQuestions := make([]*model.QuestionBankQuestion, 0)
	for _, bank := range banks {
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
