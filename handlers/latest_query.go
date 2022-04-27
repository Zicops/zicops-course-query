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

func LatestCourses(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, status *model.Status) (*model.PaginatedCourse, error) {
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
	var statusNew model.Status
	if status == nil {
		statusNew = model.StatusPublished
	} else {
		statusNew = *status
	}
	qryStr := fmt.Sprintf(`SELECT * from coursez.course where status='%s' and updated_at <= %d  ALLOW FILTERING`, statusNew, *publishTime)
	getCourses := func(page []byte) (courses []coursez.Course, nextPage []byte, err error) {
		q := global.CassSession.Session.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return courses, iter.PageState(), iter.Select(&courses)
	}
	courses, newPage, err := getCourses(newPage)
	log.Println(len(courses))
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
	var outputResponse model.PaginatedCourse
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to upload image to course: %v", err.Error())
		return nil, err
	}
	allCourses := make([]*model.Course, 0)
	for _, copiedCourse := range courses {
		course := copiedCourse
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
		allCourses = append(allCourses, &currentCourse)
	}
	outputResponse.Courses = allCourses
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}
