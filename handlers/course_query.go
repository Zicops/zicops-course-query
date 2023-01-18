package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetCourseByID(ctx context.Context, courseID []*string) ([]*model.Course, error) {
	//set the query, after getting results iterate over them using goroutines
	course := &coursez.Course{
		ID:                 "",
		Name:               "",
		Description:        "",
		Summary:            "",
		Instructor:         "",
		ImageBucket:        "",
		Image:              "",
		PreviewVideoBucket: "",
		PreviewVideo:       "",
		TileImageBucket:    "",
		TileImage:          "",
		Owner:              "",
		Duration:           0,
		ExpertiseLevel:     "",
		Language:           []string{},
		Benefits:           []string{},
		Outcomes:           []string{},
		CreatedAt:          0,
		UpdatedAt:          0,
		Type:               "",
		Prequisites:        []string{},
		GoodFor:            []string{},
		MustFor:            []string{},
		RelatedSkills:      []string{},
		PublishDate:        "",
		ExpiryDate:         "",
		QARequired:         false,
		Approvers:          []string{},
		CreatedBy:          "",
		UpdatedBy:          "",
		Status:             "",
		IsActive:           false,
		IsDisplay:          false,
		ExpectedCompletion: "",
		Category:           "",
		SubCategory:        "",
		SubCategories:      []coursez.SubCat{},
		LspId:              "",
		Publisher:          "",
	}
	key := "GetCourseByID" + fmt.Sprintf("%v", courseID)
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	//lspID := claims["lsp_id"].(string)
	result, err := redis.GetRedisValue(key)
	if err != nil {
		log.Error("Error in getting redis value for key: ", key)
	}
	if result != "" {
		err = json.Unmarshal([]byte(result), &course)
		if err != nil {
			log.Error("Error in unmarshalling redis value for key: ", key)
		}
	}
	//from here we will write query if our cache value is nil
	res := make([]*model.Course, len(courseID))
	var wg sync.WaitGroup
	if course.ID == "" || role == "admin" {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session

		//iterate through each courseId value and store it in res variable
		for i, vv := range courseID {
			//this is to avoid pointers overlapping
			v := vv

			wg.Add(1)
			//send each value to different go routines and store the result
			go func(v *string, i int) {
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
				}
				if len(courses) <= 0 {
					log.Errorf("course not found: %v", err)
				}
				course = &courses[0]

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
				}
				tileUrl := course.TileImage
				if course.TileImageBucket != "" {
					tileUrl = storageC.GetSignedURLForObject(course.TileImageBucket)
				}
				imageUrl := course.Image
				if course.ImageBucket != "" {
					imageUrl = storageC.GetSignedURLForObject(course.ImageBucket)
				}
				previewUrl := course.PreviewVideo
				if course.PreviewVideoBucket != "" {
					previewUrl = storageC.GetSignedURLForObject(course.PreviewVideoBucket)
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
					redis.SetTTL(key, 3600)
					err = redis.SetRedisValue(key, string(redisBytes))
					if err != nil {
						log.Errorf("Failed to set redis value: %v", err.Error())
					}
				}
				res[i] = &currentCourse
				wg.Done()
			}(v, i)
			wg.Wait()
		}
	}

	return res, nil
}
