package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
)

func GetCohortCourseMaps(ctx context.Context, cohortID *string) ([]*model.CourseCohort, error) {
	key := "GetCohortCourseMaps" + *cohortID
	result, err := redis.GetRedisValue(key)
	if err != nil {
		log.Errorf("GetCohortCourseMaps: %v", err)
	}
	if result != "" {
		var resultOutput []*model.CourseCohort
		err = json.Unmarshal([]byte(result), &resultOutput)
		if err != nil {
			log.Errorf("GetCohortCourseMaps: %v", err)
		}
		return resultOutput, nil
	}

	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from coursez.course_cohort_mapping where cohortid = '%s'  ALLOW FILTERING`, *cohortID)
	getCCohorts := func() (banks []coursez.CourseCohortMapping, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	cCohorts, err := getCCohorts()
	if err != nil {
		return nil, err
	}
	allSections := make([]*model.CourseCohort, 0)
	for _, cohort := range cCohorts {
		input := cohort
		created := strconv.FormatInt(input.CreatedAt, 10)
		updated := strconv.FormatInt(input.UpdatedAt, 10)
		currentQuestion := &model.CourseCohort{
			ID:           &input.ID,
			CourseID:     &input.CourseID,
			CourseType:   &input.CourseType,
			CohortID:     &input.CohortID,
			CourseStatus: &input.CourseStatus,
			LspID:        &input.LspID,
			IsMandatory:  &input.IsMandatory,
			AddedBy:      &input.AddedBy,
			IsActive:     &input.IsActive,
			CreatedBy:    &input.CreatedBy,
			UpdatedBy:    &input.UpdatedBy,
			CreatedAt:    &created,
			UpdatedAt:    &updated,
			CohortCode:   &input.CohortCode,
		}
		allSections = append(allSections, currentQuestion)
	}

	redisBytes, err := json.Marshal(allSections)
	if err != nil {
		log.Errorf("GetCohortCourseMaps: %v", err)
	} else {
		redis.SetTTL(key, 3600)
		err = redis.SetRedisValue(key, string(redisBytes))
		if err != nil {
			log.Errorf("GetCohortCourseMaps: %v", err)
		}
	}
	return allSections, nil
}
