package handlers

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetCourseByID(ctx context.Context, courseID *string) (*model.Course, error) {
	course := &coursez.Course{}
	qryStr := fmt.Sprintf(`SELECT * from coursez.course where id='%s'`, *courseID)
	err := global.CassSession.Session.Query(qryStr, nil).Scan(course)
	if err != nil {
		return nil, err
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
		language = append(language, &lang)
	}
	for _, take := range course.Benefits {
		takeaways = append(takeaways, &take)
	}
	for _, out := range course.Outcomes {
		outcomes = append(outcomes, &out)
	}
	for _, preq := range course.Prequisites {
		prequisites = append(prequisites, &preq)
	}
	for _, good := range course.GoodFor {
		goodFor = append(goodFor, &good)
	}
	for _, must := range course.MustFor {
		mustFor = append(mustFor, &must)
	}
	for _, relSkill := range course.RelatedSkills {
		relatedSkills = append(relatedSkills, &relSkill)
	}
	for _, approver := range course.Approvers {
		approvers = append(approvers, &approver)
	}
	for _, subCat := range course.SubCategories {
		var subC coursez.SubCat
		var subCR model.SubCategories
		subC.Name = subCat.Name
		subC.Rank = subCat.Rank
		subCR.Name = &subCat.Name
		subCR.Rank = &subCat.Rank
		subCatsRes = append(subCatsRes, &subCR)
	}

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to initialize storage: %v", err.Error())
		return nil, err
	}
	tileUrl := storageC.GetSignedURLForObject(course.TileImageBucket)
	imageUrl := storageC.GetSignedURLForObject(course.ImageBucket)
	previewUrl := storageC.GetSignedURLForObject(course.PreviewVideoBucket)
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
	return &currentCourse, nil
}
