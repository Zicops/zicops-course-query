package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetCourseByID(ctx context.Context, courseID []*string) ([]*model.Course, error) {

	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//from here we will write query if our cache value is nil
	res := make([]*model.Course, len(courseID))
	var wg sync.WaitGroup
	{
		session, err := global.CassPool.GetSession(ctx, "coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		//iterate through each courseId value and store it in res variable
		for i, vv := range courseID {
			//this is to avoid pointers overlapping
			if vv == nil {
				continue
			}
			var courseI coursez.Course
			vvv := vv
			key := fmt.Sprintf("course:%s", *vvv)
			{
				result, err := redis.GetRedisValue(ctx, key)
				if err != nil {
					log.Error("Error in getting redis value for key: ", key)
				}
				if result != "" {
					err = json.Unmarshal([]byte(result), &courseI)
					if err != nil {
						log.Error("Error in unmarshalling redis value for key: ", key)
					}
				}
			}
			wg.Add(1)
			//send each value to different go routines and store the result
			go func(v *string, i int, cc coursez.Course) {
				defer wg.Done()
				var course coursez.Course
				if cc.ID != "" {
					course = cc
				} else {
					qryStr := fmt.Sprintf(`SELECT * from coursez.course where id='%s' ALLOW FILTERING`, *v)
					getCourse := func() (courses []coursez.Course, err error) {
						q := CassSession.Query(qryStr, nil)
						defer q.Release()
						iter := q.Iter()
						return courses, iter.Select(&courses)
					}
					courses, err := getCourse()
					if err != nil {
						log.Errorf("%v", err)
						return
					}
					if len(courses) <= 0 {
						log.Errorf("course not found: %v", err)
						return
					}
					course = courses[0]
				}
				createdAt := strconv.FormatInt(course.CreatedAt, 10)
				updatedAt := strconv.FormatInt(course.UpdatedAt, 10)
				language := make([]*string, 0)
				takeaways := make([]*string, 0)
				outcomes := make([]*string, 0)
				prequisites := make([]*string, 0)
				goodFor := make([]*string, 0)
				mustFor := make([]*string, 0)
				relatedSkills := make([]*string, 0)
				approvers := make([]*string, 0)
				subCatsRes := make([]*model.SubCategories, 0)

				for _, lang := range course.Language {
					langCopied := lang
					language = append(language, &langCopied)
				}
				for _, take := range course.Benefits {
					takeCopied := take
					takeaways = append(takeaways, &takeCopied)
				}
				for _, out := range course.Outcomes {
					outCopied := out
					outcomes = append(outcomes, &outCopied)
				}
				for _, preq := range course.Prequisites {
					preCopied := preq
					prequisites = append(prequisites, &preCopied)
				}
				for _, good := range course.GoodFor {
					goodCopied := good
					goodFor = append(goodFor, &goodCopied)
				}
				for _, must := range course.MustFor {
					mustCopied := must
					mustFor = append(mustFor, &mustCopied)
				}
				for _, relSkill := range course.RelatedSkills {
					relCopied := relSkill
					relatedSkills = append(relatedSkills, &relCopied)
				}
				for _, approver := range course.Approvers {
					appoverCopied := approver
					approvers = append(approvers, &appoverCopied)
				}
				for _, subCat := range course.SubCategories {
					subCopied := subCat
					var subCR model.SubCategories
					subCR.Name = &subCopied.Name
					subCR.Rank = &subCopied.Rank
					subCatsRes = append(subCatsRes, &subCR)
				}

				storageC := bucket.NewStorageHandler()
				gproject := googleprojectlib.GetGoogleProjectID()
				err = storageC.InitializeStorageClient(ctx, gproject, course.LspId)
				if err != nil {
					log.Errorf("Failed to initialize storage: %v", err.Error())
					return
				}
				tileUrl := course.TileImage
				if course.TileImageBucket != "" {
					tileUrl = storageC.GetSignedURLForObjectCache(ctx, course.TileImageBucket)
				}
				imageUrl := course.Image
				if course.ImageBucket != "" {
					imageUrl = storageC.GetSignedURLForObjectCache(ctx, course.ImageBucket)
				}
				previewUrl := course.PreviewVideo
				if course.PreviewVideoBucket != "" {
					previewUrl = storageC.GetSignedURLForObjectCache(ctx, course.PreviewVideoBucket)
				}
				var statusNew model.Status
				if course.Status == model.StatusApprovalPending.String() {
					statusNew = model.StatusApprovalPending
				} else if course.Status == model.StatusApproved.String() {
					statusNew = model.StatusApproved
				} else if course.Status == model.StatusRejected.String() {
					statusNew = model.StatusRejected
				} else if course.Status == model.StatusSaved.String() {
					statusNew = model.StatusSaved
				} else if course.Status == model.StatusOnHold.String() {
					statusNew = model.StatusOnHold
				} else if course.Status == model.StatusPublished.String() {
					statusNew = model.StatusPublished
				}
				currentCourse := model.Course{
					ID:                 &course.ID,
					Name:               &course.Name,
					LspID:              &course.LspId,
					Publisher:          &course.Publisher,
					Description:        &course.Description,
					Summary:            &course.Summary,
					Instructor:         &course.Instructor,
					Owner:              &course.Owner,
					Duration:           &course.Duration,
					ExpertiseLevel:     &course.ExpertiseLevel,
					Language:           language,
					Benefits:           takeaways,
					Outcomes:           outcomes,
					CreatedAt:          &createdAt,
					UpdatedAt:          &updatedAt,
					Type:               &course.Type,
					Prequisites:        prequisites,
					GoodFor:            goodFor,
					MustFor:            mustFor,
					RelatedSkills:      relatedSkills,
					PublishDate:        &course.PublishDate,
					ExpiryDate:         &course.ExpiryDate,
					ExpectedCompletion: &course.ExpectedCompletion,
					QaRequired:         &course.QARequired,
					Approvers:          approvers,
					CreatedBy:          &course.CreatedBy,
					UpdatedBy:          &course.UpdatedBy,
					Status:             &statusNew,
					IsDisplay:          &course.IsDisplay,
					Category:           &course.Category,
					SubCategory:        &course.SubCategory,
					SubCategories:      subCatsRes,
					IsActive:           &course.IsActive,
				}
				if course.TileImageBucket != "" {
					currentCourse.TileImage = &tileUrl
				}
				if course.ImageBucket != "" {
					currentCourse.Image = &imageUrl
				}
				if course.PreviewVideoBucket != "" {
					currentCourse.PreviewVideo = &previewUrl
				}
				redisBytes, err := json.Marshal(course)
				if err != nil {
					log.Errorf("Failed to marshal course: %v", err.Error())
				} else {
					err = redis.SetRedisValue(ctx, key, string(redisBytes))
					if err != nil {
						log.Errorf("Failed to set redis value: %v", err.Error())
					}
					err = redis.SetTTL(ctx, key, 60)
					if err != nil {
						log.Errorf("Failed to set redis ttl: %v", err.Error())
					}
				}
				res[i] = &currentCourse

			}(vvv, i, courseI)
		}
		wg.Wait()
	}

	return res, nil
}

func GetBasicCourseStats(ctx context.Context, input *model.BasicCourseStatsInput) (*model.BasicCourseStats, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	if input.Categories != nil && input.SubCategories != nil && input.ExpertiseLevel != nil && input.Languages != nil {
		return nil, fmt.Errorf("only one of the following can be provided: Categories, SubCategories, ExpertiseLevel, Languages")
	}
	whereClause := fmt.Sprintf(" WHERE lsp_id = '%s' ", input.LspID)
	if input.CourseStatus != nil {
		whereClause = fmt.Sprintf("%s AND status = '%s' ", whereClause, *input.CourseStatus)
	}
	if input.CourseType != nil {
		whereClause = fmt.Sprintf("%s AND type = '%s' ", whereClause, *input.CourseType)
	}
	if input.Duration != nil {
		whereClause = fmt.Sprintf("%s AND duration <= %d ", whereClause, *input.Duration)
	}
	if input.Owner != nil {
		whereClause = fmt.Sprintf("%s AND owner = '%s' ", whereClause, *input.Owner)
	}
	if input.CreatedBy != nil {
		whereClause = fmt.Sprintf("%s AND created_by = '%s' ", whereClause, *input.CreatedBy)
	}
	catStats := make([]*model.Count, 0)
	var wg sync.WaitGroup
	if input.Categories != nil {
		catStats = make([]*model.Count, len(input.Categories))
		for i, vv := range input.Categories {
			vvv := vv
			if vvv == nil {
				continue
			}
			wg.Add(1)
			go func(v *string, i int) {
				copiedCat := *v
				tempClause := fmt.Sprintf("%s AND category = '%s' ", whereClause, copiedCat)
				query := fmt.Sprintf("SELECT COUNT(*) FROM coursez.course %s ALLOW FILTERING", tempClause)
				iter := CassUserSession.Query(query, nil).Iter()
				count := 0
				iter.Scan(&count)
				currentStat := model.Count{
					Name:  copiedCat,
					Count: count,
				}
				catStats[i] = &currentStat
				wg.Done()
			}(vvv, i)
		}
	}
	subCatStats := make([]*model.Count, 0)
	if input.SubCategories != nil {
		subCatStats = make([]*model.Count, len(input.SubCategories))
		for i, vv := range input.SubCategories {
			vvv := vv
			if vvv == nil {
				continue
			}
			wg.Add(1)
			go func(v *string, i int) {
				copiedSubCat := *v
				tempClause := fmt.Sprintf("%s AND sub_category = '%s' ", whereClause, copiedSubCat)
				query := fmt.Sprintf("SELECT COUNT(*) FROM coursez.course %s ALLOW FILTERING", tempClause)
				iter := CassUserSession.Query(query, nil).Iter()
				count := 0
				iter.Scan(&count)
				currentStat := model.Count{
					Name:  copiedSubCat,
					Count: count,
				}
				subCatStats[i] = &currentStat
				wg.Done()
			}(vvv, i)
		}
	}
	expertiseStats := make([]*model.Count, 0)
	if input.ExpertiseLevel != nil {
		expertiseStats = make([]*model.Count, len(input.ExpertiseLevel))
		for i, vv := range input.ExpertiseLevel {
			vvv := vv
			if vvv == nil {
				continue
			}
			wg.Add(1)
			go func(v *string, i int) {
				copiedExpertise := *v
				tempClause := fmt.Sprintf("%s AND expertise_level = '%s' ", whereClause, copiedExpertise)
				query := fmt.Sprintf("SELECT COUNT(*) FROM coursez.course %s ALLOW FILTERING", tempClause)
				iter := CassUserSession.Query(query, nil).Iter()
				count := 0
				iter.Scan(&count)
				currentStat := model.Count{
					Name:  copiedExpertise,
					Count: count,
				}
				expertiseStats[i] = &currentStat
				wg.Done()
			}(vvv, i)
		}
	}
	languageStats := make([]*model.Count, 0)
	if input.Languages != nil {
		languageStats = make([]*model.Count, len(input.Languages))
		for i, vv := range input.Languages {
			vvv := vv
			if vvv == nil {
				continue
			}
			wg.Add(1)
			go func(v *string, i int) {
				copiedLanguage := *v
				tempClause := fmt.Sprintf("%s AND language CONTAINS '%s' ", whereClause, copiedLanguage)
				query := fmt.Sprintf("SELECT COUNT(*) FROM coursez.course %s ALLOW FILTERING", tempClause)
				iter := CassUserSession.Query(query, nil).Iter()
				count := 0
				iter.Scan(&count)
				currentStat := model.Count{
					Name:  copiedLanguage,
					Count: count,
				}
				languageStats[i] = &currentStat
				wg.Done()
			}(vvv, i)
		}
	}
	wg.Wait()
	res := model.BasicCourseStats{
		CourseStatus:   input.CourseStatus,
		CourseType:     input.CourseType,
		Duration:       input.Duration,
		Owner:          input.Owner,
		CreatedBy:      input.CreatedBy,
		LspID:          input.LspID,
		Categories:     catStats,
		SubCategories:  subCatStats,
		ExpertiseLevel: expertiseStats,
		Languages:      languageStats,
	}
	return &res, nil
}

func GetCourseCountStats(ctx context.Context, lspID *string, status string, typeArg string) (*model.CourseCountStats, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lsp := claims["lsp_id"].(string)
	if lspID != nil {
		lsp = *lspID
	}
	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	qryStr := fmt.Sprintf(`SELECT COUNT(*) FROM coursez.course WHERE lsp_id='%s' AND status='%s' AND type='%s' ALLOW FILTERING`, lsp, status, typeArg)
	iter := CassUserSession.Query(qryStr, nil).Iter()
	count := 0
	iter.Scan(&count)

	//we have count in count variable
	res := model.CourseCountStats{
		LspID:        &lsp,
		CourseStatus: &status,
		CourseType:   &typeArg,
		Count:        &count,
	}
	return &res, nil
}
