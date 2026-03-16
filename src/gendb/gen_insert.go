package gendb

import (
	"log/slog"

	"github.com/shabs76/roro-local-server/constants"
)

func SaveGeneral(sq string, vals []any) *constants.AnswerState {
	db, er := InitDb()

	if er != nil {
		slog.Error(er.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  er.Error(),
			Adv:   "none",
		}
	}

	defer db.Close()

	stmt, err := db.Prepare(sq)

	if err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to create statement for this request. Please try again",
			Adv:   "none",
		}
	}

	_, erE := stmt.Exec(vals...)
	if erE != nil {
		slog.Error(erE.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Statement execution has failed. Please try again",
			Adv:   "none",
		}
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}
}
