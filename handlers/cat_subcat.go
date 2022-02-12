package handlers

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-course-query/global"
)

func GetCategories(ctx context.Context) ([]*string, error) {
	log.Info("GetCategories")
	catergories := make([]string, 0)
	resultOutput := make([]*string, 0)
	getQueryCassandra := global.CassSession.Session.Query(coursez.CatTable.Get()).Bind(catergories)
	if err := getQueryCassandra.ExecRelease(); err != nil {
		return nil, err
	}
	for _, category := range catergories {
		resultOutput = append(resultOutput, &category)
	}
	return resultOutput, nil
}

func GetSubCategories(ctx context.Context) ([]*string, error) {
	log.Info("GetCategories")
	subCatergories := make([]string, 0)
	resultOutput := make([]*string, 0)
	getQueryCassandra := global.CassSession.Session.Query(coursez.SubCatTable.Get()).Bind(subCatergories)
	if err := getQueryCassandra.ExecRelease(); err != nil {
		return nil, err
	}
	for _, category := range subCatergories {
		resultOutput = append(resultOutput, &category)
	}
	return resultOutput, nil
}
