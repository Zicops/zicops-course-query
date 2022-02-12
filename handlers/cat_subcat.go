package handlers

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-course-query/global"
)

func GetCategories(ctx context.Context) ([]*string, error) {
	log.Info("GetCategories")

	resultOutput := make([]*string, 0)
	getQueryCassandra := global.CassSession.Session.Query("SELECT * from coursez.category", nil)

	iter := getQueryCassandra.Iter()
	var tempCat string
	for iter.Scan(&tempCat) {
		copyCat := tempCat
		resultOutput = append(resultOutput, &copyCat)
	}
	err := iter.Close()
	if err != nil {
		return resultOutput, err
	}
	return resultOutput, nil
}

func GetSubCategories(ctx context.Context) ([]*string, error) {
	log.Info("GetSubCategories")

	resultOutput := make([]*string, 0)
	getQueryCassandra := global.CassSession.Session.Query("SELECT * from coursez.sub_category", nil)

	iter := getQueryCassandra.Iter()
	var tempCat string
	for iter.Scan(&tempCat) {
		copyCat := tempCat
		resultOutput = append(resultOutput, &copyCat)
	}
	err := iter.Close()
	if err != nil {
		return resultOutput, err
	}
	return resultOutput, nil
}
