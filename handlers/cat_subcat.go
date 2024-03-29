package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

func GetCategories(ctx context.Context) ([]*string, error) {
	log.Info("GetCategories")
	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	resultOutput := make([]*string, 0)
	getQueryCassandra := CassSession.Query("SELECT * from coursez.category", nil)

	iter := getQueryCassandra.Iter()
	var tempCat string
	for iter.Scan(&tempCat) {
		copyCat := tempCat
		resultOutput = append(resultOutput, &copyCat)
	}
	err = iter.Close()
	if err != nil {
		return resultOutput, err
	}
	return resultOutput, nil
}

func GetSubCategories(ctx context.Context) ([]*string, error) {
	log.Info("GetSubCategories")
	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	resultOutput := make([]*string, 0)
	getQueryCassandra := CassSession.Query("SELECT * from coursez.sub_category", nil)

	iter := getQueryCassandra.Iter()
	var tempCat string
	for iter.Scan(&tempCat) {
		copyCat := tempCat
		resultOutput = append(resultOutput, &copyCat)
	}
	err = iter.Close()
	if err != nil {
		return resultOutput, err
	}
	return resultOutput, nil
}

func GetSubCategoriesForSub(ctx context.Context, cat *string) ([]*string, error) {
	log.Info("GetSubCategoriesForSub")
	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	resultOutput := make([]*string, 0)
	qryStr := fmt.Sprintf(`SELECT sub_category from coursez.cat_sub_mapping WHERE category = '%s'  ALLOW FILTERING`, *cat)
	getQueryCassandra := CassSession.Query(qryStr, nil)

	iter := getQueryCassandra.Iter()
	var tempCat string
	for iter.Scan(&tempCat) {
		copyCat := tempCat
		resultOutput = append(resultOutput, &copyCat)
	}
	err = iter.Close()
	if err != nil {
		return resultOutput, err
	}
	return resultOutput, nil
}

func AllCatMain(ctx context.Context, lspIds []*string, searchText *string) ([]*model.CatMain, error) {
	log.Info("AllCatMain")
	key := "zicops_all_cat_main"
	if len(lspIds) > 0 {
		for _, lspId := range lspIds {
			if lspId != nil {
				lc := strings.ToLower(*lspId)
				key = key + lc
			}
		}
	}
	if searchText != nil && *searchText != "" {
		key = key + *searchText
	}
	key = base64.StdEncoding.EncodeToString([]byte(key))
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	log.Infof("claims: %v", claims)
	role := strings.ToLower(claims["role"].(string))
	cats := make([]coursez.CatMain, 0)
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil && result != "" {
		log.Info("Got value from redis")
		err = json.Unmarshal([]byte(result), &cats)
		if err != nil {
			log.Errorf("Failed to unmarshal value from redis: %v", err.Error())
		}
	}
	if len(cats) <= 0 || role != "learner" {
		cats = make([]coursez.CatMain, 0)
		session, err := global.CassPool.GetSession(ctx, "coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session
		whereClause := "WHERE "
		if len(lspIds) > 0 {
			// cassandra contains clauses using lspIds
			for i, lspId := range lspIds {
				if lspId == nil || *lspId == "" {
					continue
				}
				if i == 0 || whereClause == "WHERE " {
					whereClause = whereClause + " lsps CONTAINS '" + *lspId + "'"
				} else {
					whereClause = whereClause + " AND lsps CONTAINS '" + *lspId + "'"
				}
			}
			if searchText != nil && *searchText != "" {
				searchTextLower := strings.ToLower(*searchText)
				words := strings.Split(searchTextLower, " ")
				for _, word := range words {
					whereClause = whereClause + " AND  words CONTAINS '" + word + "'"
				}
			}

		} else {
			if searchText != nil && *searchText != "" {
				searchTextLower := strings.ToLower(*searchText)
				words := strings.Split(searchTextLower, " ")
				for i, word := range words {
					if i == 0 {
						whereClause = whereClause + "words CONTAINS '" + word + "'"
					} else {
						whereClause = whereClause + " AND  words CONTAINS '" + word + "'"
					}
				}
			}
		}
		qryStr := `SELECT * from coursez.cat_main ` + whereClause + ` ALLOW FILTERING`
		getCats := func() (banks []coursez.CatMain, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return banks, iter.Select(&banks)
		}
		cats, err = getCats()
		if err != nil {
			return nil, err
		}
	}
	resultOutput := make([]*model.CatMain, len(cats))
	if len(cats) <= 0 {
		return resultOutput, nil
	}
	var wg sync.WaitGroup
	for i, cat := range cats {
		cc := cat
		wg.Add(1)
		go func(i int, copiedCat coursez.CatMain) {
			createdAt := strconv.FormatInt(copiedCat.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedCat.UpdatedAt, 10)
			imageUrl := copiedCat.ImageURL
			if copiedCat.ImageBucket != "" {
				storageC := bucket.NewStorageHandler()
				gproject := googleprojectlib.GetGoogleProjectID()
				err = storageC.InitializeStorageClient(ctx, gproject, "coursez-catimages")
				if err != nil {
					log.Errorf("Failed to initialize storage: %v", err.Error())
					return
				}
				imageUrl = storageC.GetSignedURLForObjectCache(ctx, copiedCat.ImageBucket)
			}
			currentCat := model.CatMain{
				ID:          &copiedCat.ID,
				Name:        &copiedCat.Name,
				Description: &copiedCat.Description,
				Code:        &copiedCat.Code,
				ImageURL:    &imageUrl,
				CreatedBy:   &copiedCat.CreatedBy,
				CreatedAt:   &createdAt,
				UpdatedAt:   &updatedAt,
				UpdatedBy:   &copiedCat.UpdatedBy,
				IsActive:    &copiedCat.IsActive,
			}
			resultOutput[i] = &currentCat
			wg.Done()
		}(i, cc)
	}
	wg.Wait()
	redisValue, err := json.Marshal(cats)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisValue))
		redis.SetTTL(ctx, key, 3600)
	}
	return resultOutput, nil
}

func AllSubCatMain(ctx context.Context, lspIds []*string, searchText *string) ([]*model.SubCatMain, error) {
	log.Info("AllSubCatMain")
	key := "zicops_all_subcat_main"
	if len(lspIds) > 0 {
		for _, lspId := range lspIds {
			if lspId != nil {
				lc := strings.ToLower(*lspId)
				key = key + lc
			}

		}
	}
	if searchText != nil && *searchText != "" {
		key = key + *searchText
	}
	key = base64.StdEncoding.EncodeToString([]byte(key))
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	cats := make([]coursez.SubCatMain, 0)
	result, err := redis.GetRedisValue(ctx, key)
	if err != nil {
		log.Errorf("Failed to get value from redis: %v", err.Error())
	}
	if result != "" {
		log.Info("Got value from redis")
		err = json.Unmarshal([]byte(result), &cats)
		if err != nil {
			log.Errorf("Failed to unmarshal value from redis: %v", err.Error())
		}
	}
	if len(cats) <= 0 || role != "learner" {
		cats = make([]coursez.SubCatMain, 0)
		session, err := global.CassPool.GetSession(ctx, "coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session
		whereClause := "WHERE "
		if len(lspIds) > 0 {
			// cassandra contains clauses using lspIds
			for i, lspId := range lspIds {
				if lspId == nil || *lspId == "" {
					continue
				}
				if i == 0 || whereClause == "WHERE " {
					whereClause = whereClause + " lsps CONTAINS '" + *lspId + "'"
				} else {
					whereClause = whereClause + " AND lsps CONTAINS '" + *lspId + "'"
				}
			}
			if searchText != nil && *searchText != "" {
				searchTextLower := strings.ToLower(*searchText)
				words := strings.Split(searchTextLower, " ")
				for _, word := range words {
					whereClause = whereClause + " AND  words CONTAINS '" + word + "'"
				}
			}

		} else {
			if searchText != nil && *searchText != "" {
				searchTextLower := strings.ToLower(*searchText)
				words := strings.Split(searchTextLower, " ")
				for i, word := range words {
					if i == 0 {
						whereClause = whereClause + "words CONTAINS '" + word + "'"
					} else {
						whereClause = whereClause + " AND  words CONTAINS '" + word + "'"
					}
				}
			}
		}
		qryStr := `SELECT * from coursez.sub_cat_main ` + whereClause + ` ALLOW FILTERING`
		getCats := func() (banks []coursez.SubCatMain, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return banks, iter.Select(&banks)
		}
		cats, err = getCats()
		if err != nil {
			return nil, err
		}
	}
	resultOutput := make([]*model.SubCatMain, len(cats))
	if len(cats) <= 0 {
		return resultOutput, nil
	}

	var wg sync.WaitGroup
	for i, cat := range cats {
		cc := cat
		wg.Add(1)
		go func(i int, copiedCat coursez.SubCatMain) {
			createdAt := strconv.FormatInt(copiedCat.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedCat.UpdatedAt, 10)
			imageUrl := copiedCat.ImageURL
			if copiedCat.ImageBucket != "" {
				storageC := bucket.NewStorageHandler()
				gproject := googleprojectlib.GetGoogleProjectID()
				err = storageC.InitializeStorageClient(ctx, gproject, "coursez-catimages")
				if err != nil {
					log.Errorf("Failed to initialize storage: %v", err.Error())
					return
				}
				imageUrl = storageC.GetSignedURLForObjectCache(ctx, copiedCat.ImageBucket)
			}
			currentCat := model.SubCatMain{
				ID:          &copiedCat.ID,
				Name:        &copiedCat.Name,
				Description: &copiedCat.Description,
				Code:        &copiedCat.Code,
				ImageURL:    &imageUrl,
				CreatedBy:   &copiedCat.CreatedBy,
				CreatedAt:   &createdAt,
				UpdatedAt:   &updatedAt,
				UpdatedBy:   &copiedCat.UpdatedBy,
				IsActive:    &copiedCat.IsActive,
				CatID:       &copiedCat.ParentID,
			}
			resultOutput[i] = &currentCat
			wg.Done()
		}(i, cc)
	}
	wg.Wait()
	redisValue, err := json.Marshal(cats)
	if err == nil {
		err = redis.SetRedisValue(ctx, key, string(redisValue))
		if err != nil {
			log.Errorf("Failed to set value in redis: %v", err.Error())
		}
		redis.SetTTL(ctx, key, 3600)
	}
	return resultOutput, nil
}

func AllSubCatByCatID(ctx context.Context, catID *string) ([]*model.SubCatMain, error) {
	log.Info("AllSubCatByCatID")
	key := "AllSubCatByCatID" + *catID
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	role := strings.ToLower(claims["role"].(string))
	cats := make([]coursez.SubCatMain, 0)
	result, err := redis.GetRedisValue(ctx, key)
	if err != nil {
		log.Errorf("Failed to get value from redis: %v", err.Error())
	}
	if result != "" {
		log.Info("Got value from redis")
		err = json.Unmarshal([]byte(result), &cats)
		if err != nil {
			log.Errorf("Failed to unmarshal value from redis: %v", err.Error())
		}
	}
	if len(cats) <= 0 || role != "learner" {
		session, err := global.CassPool.GetSession(ctx, "coursez")
		if err != nil {
			return nil, err
		}
		CassSession := session
		qryStr := fmt.Sprintf(`SELECT * from coursez.sub_cat_main WHERE parent_id = '%s' ALLOW FILTERING`, *catID)
		getCats := func() (banks []coursez.SubCatMain, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return banks, iter.Select(&banks)
		}
		cats, err = getCats()
		if err != nil {
			return nil, err
		}
	}
	resultOutput := make([]*model.SubCatMain, len(cats))
	if len(cats) <= 0 {
		return resultOutput, nil
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject, "coursez-catimages")
	if err != nil {
		log.Errorf("Failed to initialize storage: %v", err.Error())
		return nil, err
	}
	for i, cat := range cats {
		copiedCat := cat
		createdAt := strconv.FormatInt(copiedCat.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedCat.UpdatedAt, 10)
		imageUrl := copiedCat.ImageURL
		if copiedCat.ImageBucket != "" {
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject, "coursez-catimages")
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
			}
			imageUrl = storageC.GetSignedURLForObjectCache(ctx, copiedCat.ImageBucket)
		}
		currentCat := model.SubCatMain{
			ID:          &copiedCat.ID,
			Name:        &copiedCat.Name,
			Description: &copiedCat.Description,
			Code:        &copiedCat.Code,
			ImageURL:    &imageUrl,
			CreatedBy:   &copiedCat.CreatedBy,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			UpdatedBy:   &copiedCat.UpdatedBy,
			IsActive:    &copiedCat.IsActive,
			CatID:       &copiedCat.ParentID,
		}
		resultOutput[i] = &currentCat
	}
	redisValue, err := json.Marshal(cats)
	if err == nil {
		err = redis.SetRedisValue(ctx, key, string(redisValue))
		if err != nil {
			log.Errorf("Failed to set value in redis: %v", err.Error())
		}
		redis.SetTTL(ctx, key, 60)
	}
	return resultOutput, nil
}
