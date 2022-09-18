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

func GetCategories(ctx context.Context) ([]*string, error) {
	log.Info("GetCategories")
	session, err := cassandra.GetCassSession("coursez")
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
	session, err := cassandra.GetCassSession("coursez")
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
	session, err := cassandra.GetCassSession("coursez")
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

func AllCatMain(ctx context.Context) ([]*model.CatMain, error) {
	log.Info("AllCatMain")
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := `SELECT * from coursez.cat_main `
	getCats := func() (banks []coursez.CatMain, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	cats, err := getCats()
	if err != nil {
		return nil, err
	}
	resultOutput := make([]*model.CatMain, 0)
	for _, cat := range cats {
		copiedCat := cat
		createdAt := strconv.FormatInt(copiedCat.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedCat.UpdatedAt, 10)
		imageUrl := copiedCat.ImageURL
		if copiedCat.ImageBucket != "" {
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
				continue
			}
			imageUrl = storageC.GetSignedURLForObject(copiedCat.ImageBucket)
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
		resultOutput = append(resultOutput, &currentCat)

	}
	return resultOutput, nil
}

func AllSubCatMain(ctx context.Context) ([]*model.SubCatMain, error) {
	log.Info("AllSubCatMain")
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := `SELECT * from coursez.sub_cat_main `
	getCats := func() (banks []coursez.SubCatMain, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return banks, iter.Select(&banks)
	}
	cats, err := getCats()
	if err != nil {
		return nil, err
	}
	resultOutput := make([]*model.SubCatMain, 0)
	for _, cat := range cats {
		copiedCat := cat
		createdAt := strconv.FormatInt(copiedCat.CreatedAt, 10)
		updatedAt := strconv.FormatInt(copiedCat.UpdatedAt, 10)
		imageUrl := copiedCat.ImageURL
		if copiedCat.ImageBucket != "" {
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Errorf("Failed to initialize storage: %v", err.Error())
				continue
			}
			imageUrl = storageC.GetSignedURLForObject(copiedCat.ImageBucket)
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
		resultOutput = append(resultOutput, &currentCat)

	}
	return resultOutput, nil
}
