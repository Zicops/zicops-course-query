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
		tileUrl := storageC.GetSignedURLForObject(course.TileImageBucket)
		imageUrl := storageC.GetSignedURLForObject(course.ImageBucket)
		previewUrl := storageC.GetSignedURLForObject(course.PreviewVideoBucket)
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
