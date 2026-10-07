package gendb

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math"

	"github.com/shabs76/roro-local-server/constants"
)

func SelectGeneral(sq string, vals []any) (*constants.AnswerState, *sql.Rows) {
	db, er := InitDb()
	if er != nil {
		slog.Error(er.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  er.Error(),
			Adv:   "none",
		}, nil
	}

	res, err := db.Query(sq, vals...)
	if err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to execute search on the request. Please try again",
			Adv:   "none",
		}, nil
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, res
}

func PagenationSelect(query string, pgn, perPage int, vals []any) (*constants.AnswerState, *constants.PageNationSelect) {
	db, er := InitDb()
	if er != nil {
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  er.Error(),
			Adv:   "none",
		}, nil
	}

	res, err := db.Query(query, vals...)
	if err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to execute number search",
			Adv:   "none",
		}, nil
	}
	defer res.Close()

	type numTy struct {
		ResNum int
	}

	var num numTy
	for res.Next() {
		erS := res.Scan(&num.ResNum)
		if erS != nil {
			slog.Error(erS.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to scan number results",
				Adv:   "none",
			}, nil
		}
	}

	// calculate number base on the num per page
	pages := 1
	totalPages := int(math.Ceil(float64(num.ResNum) / float64(perPage)))
	if num.ResNum > perPage {
		pages = totalPages
	}
	// set state and limit
	var rez constants.PageNationSelect
	rez.ResNum = num.ResNum
	rez.TotalPages = totalPages
	if pgn < 1 {
		pgn = 1
	} else if pgn > pages {
		pgn = pages
		rez.State = "end"
	} else {
		rez.State = "cont"
	}

	rez.Limit = " LIMIT " + fmt.Sprint((pgn-1)*perPage) + " , " + fmt.Sprint(perPage)

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, &rez
}
