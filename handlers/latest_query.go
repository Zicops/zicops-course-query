package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/bradhe/stopwatch"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"github.com/zicops/zicops-course-query/lib/db/bucket"
	"github.com/zicops/zicops-course-query/lib/googleprojectlib"
)

func LatestCourses(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, status *model.Status, filters *model.CoursesFilters) (*model.PaginatedCourse, error) {
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
	// stringify filters
	var filtersStr string
	if filters != nil {
		/// iterate over filets and add to string which are not nil
		filtersStr = "filtered"
		if filters.Language != nil {
			filtersStr = fmt.Sprintf("%v_languages_%v", filtersStr, *filters.Language)
		}
		if filters.LspID != nil {
			filtersStr = fmt.Sprintf("%v_levels_%v", filtersStr, *filters.LspID)
		}
		if filters.Category != nil {
			filtersStr = fmt.Sprintf("%v_categories_%v", filtersStr, *filters.Category)
		}
		if filters.Owner != nil {
			filtersStr = fmt.Sprintf("%v_owner_%v", filtersStr, *filters.Owner)
		}
		if filters.Publisher != nil {
			filtersStr = fmt.Sprintf("%v_publisher_%v", filtersStr, *filters.Publisher)
		}
		if filters.SubCategory != nil {
			filtersStr = fmt.Sprintf("%v_subcategories_%v", filtersStr, *filters.SubCategory)
		}
		if filters.SearchText != nil {
			filtersStr = fmt.Sprintf("%v_search_%v", filtersStr, *filters.SearchText)
		}
		if filters.Type != nil {
			filtersStr = fmt.Sprintf("%v_types_%v", filtersStr, *filters.Type)
		}

	} else {
		filtersStr = "landed"
	}
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	role := strings.ToLower(claims["role"].(string))
	key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("latest_courses_%v_%v", role, filtersStr)))
	result, err := redis.GetRedisValue(ctx, key)
	if err != nil {
		log.Errorf("Error in getting redis value for key %v : %v", key, err)
	}
	dbCourses := make([]coursez.Course, 0)
	if result != "" {
		log.Infof("Redis value found for key: %v", key)
		err = json.Unmarshal([]byte(result), &dbCourses)
		if err != nil {
			log.Errorf("Error in unmarshalling redis value: %v", err)
		}
	}
	var newCursor string
	var statusNew model.Status
	if status == nil {
		statusNew = model.StatusPublished
	} else {
		statusNew = *status
	}
	if len(dbCourses) <= 0 || role != "learner" {
		session, err := cassandra.GetCassSession("coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session
		whereClause := fmt.Sprintf(`where status='%s' and created_at <= %d`, statusNew, *publishTime)
		if filters != nil {
			if filters.Category != nil {
				whereClause = whereClause + fmt.Sprintf(` and category='%s'`, *filters.Category)
			}
			if filters.SubCategory != nil {
				whereClause = whereClause + fmt.Sprintf(` and sub_category='%s'`, *filters.SubCategory)
			}
			if filters.Language != nil {
				whereClause = whereClause + fmt.Sprintf(` and language CONTAINS '%s'`, *filters.Language)
			}
			if filters.LspID != nil {
				whereClause = whereClause + fmt.Sprintf(` and lsp_id='%s'`, *filters.LspID)
			}
			if filters.DurationMax != nil {
				whereClause = whereClause + fmt.Sprintf(` and duration<=%d`, *filters.DurationMax)
			}
			if filters.DurationMin != nil {
				whereClause = whereClause + fmt.Sprintf(` and duration>=%d`, *filters.DurationMin)
			}
			if filters.Type != nil {
				whereClause = whereClause + fmt.Sprintf(` and type='%s'`, *filters.Type)
			}
			if filters.Owner != nil {
				whereClause = whereClause + fmt.Sprintf(` and owner='%s'`, *filters.Owner)
			}
			if filters.Publisher != nil {
				whereClause = whereClause + fmt.Sprintf(` and publisher='%s'`, *filters.Publisher)
			}
			if filters.SearchText != nil {
				if *filters.SearchText != "" {
					searchTextLower := strings.ToLower(*filters.SearchText)
					words := strings.Split(searchTextLower, " ")
					for _, word := range words {
						whereClause = whereClause + " AND  words CONTAINS '" + word + "'"
					}
				}
			}
		}
		whereClause = whereClause + " AND is_active=true"
		qryStr := fmt.Sprintf(`SELECT * from coursez.course %s ALLOW FILTERING`, whereClause)
		getCourses := func(page []byte) (courses []coursez.Course, nextPage []byte, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			q.PageState(page)
			q.PageSize(pageSizeInt)

			iter := q.Iter()
			return courses, iter.PageState(), iter.Select(&courses)
		}
		dbCourses, newPage, err = getCourses(newPage)
		if err != nil {
			return nil, err
		}
		redisBytes, err := json.Marshal(dbCourses)
		if err == nil {
			err = redis.SetRedisValue(ctx, key, string(redisBytes))
			if err != nil {
				log.Errorf("Failed to set redis value: %v", err.Error())
			}
			err = redis.SetTTL(ctx, key, 600)
			if err != nil {
				log.Errorf("Failed to set redis value: %v", err.Error())
			}
		} else {
			log.Errorf("Failed to marshal redis value: %v", err.Error())
		}
	}
	start := stopwatch.Start()
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Courses: %v", string(newCursor))

	}
	end := start.Stop()
	log.Infof("Time taken to encrypt cursor: %v", end)
	var outputResponse model.PaginatedCourse
	allCourses := make([]*model.Course, len(dbCourses))
	if len(dbCourses) <= 0 {
		outputResponse.Courses = allCourses
		return &outputResponse, nil
	}
	var wg sync.WaitGroup
	for i, cCourse := range dbCourses {
		copiedCourse := cCourse
		if copiedCourse.ID == "" {
			continue
		}
		wg.Add(1)
		go func(course coursez.Course, i int) {
			gproject := googleprojectlib.GetGoogleProjectID()
			storageC := bucket.NewStorageHandler()
			err = storageC.InitializeStorageClient(ctx, gproject, course.LspId)
			if err != nil {
				log.Errorf("Failed to initialize bucket to course: %v", err.Error())
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
			allCourses[i] = &currentCourse
			wg.Done()
		}(copiedCourse, i)
	}
	wg.Wait()
	outputResponse.Courses = allCourses
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction

	return &outputResponse, nil
}
