package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var (
	ErrQueryFailed         = errors.New("query failed")
	ErrReadResultFailed    = errors.New("failed to read result")
	ErrConvertResultFailed = errors.New("failed to convert result")
)

func DoPowerWindowQuery(ctx context.Context, qAPI api.QueryAPI, q string) (float64, error) {
	qResult, err := qAPI.Query(ctx, q)
	if err != nil {
		return 0, &multierror.Error{Errors: []error{
			ErrQueryFailed,
			err,
		}}
	}
	qResult.Next()
	if qResult.Record() == nil {
		return 0, &multierror.Error{Errors: []error{
			ErrQueryFailed,
			err,
		}}
	}
	qResultVal := qResult.Record().Value()
	_ = qResult.Close()
	if qResult.Err() != nil {
		return 0, &multierror.Error{Errors: []error{
			ErrReadResultFailed,
			qResult.Err(),
		}}
	}
	qResultValFloat, ok := qResultVal.(float64)
	if !ok {
		return 0, &multierror.Error{Errors: []error{
			ErrConvertResultFailed,
			fmt.Errorf("failed to convert '%v' to float64", qResultVal),
		}}
	}
	return qResultValFloat, nil
}

func IsQueryErrFatal(err error) bool {
	if errors.Is(err, ErrQueryFailed) || errors.Is(err, ErrReadResultFailed) {
		return false
	}
	return true
}
