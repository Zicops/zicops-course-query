package handlers

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func GetCourseByID(ctx context.Context, courseID *string) (*model.Course, error) {
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
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * from coursez.course where id='%s'`, *courseID)
	getCourse := func() (courses []coursez.Course, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return courses, iter.Select(&courses)
	}
	courses, err := getCourse()
	if err != nil {
		return nil, err
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
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to initialize storage: %v", err.Error())
		return nil, err
	}
	tileUrl := ""
	if course.TileImageBucket != "" {
		tileUrl = storageC.GetSignedURLForObject(course.TileImageBucket)
	}
	imageUrl := ""
	if course.ImageBucket != "" {
		imageUrl = storageC.GetSignedURLForObject(course.ImageBucket)
	}
	previewUrl := ""
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
	return &currentCourse, nil
}
