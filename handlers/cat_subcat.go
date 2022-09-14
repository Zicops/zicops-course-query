package handlers

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-cass-pool/cassandra"
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
