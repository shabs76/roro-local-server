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

	stmt, err := db.Prepare(sq)

	if err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to create statement for this request. Please try again",
			Adv:   "none",
		}
	}

	// The pool is shared and long-lived, so an unclosed statement stays prepared on the
	// server until its connection closes; bulk list syncs would hit
	// max_prepared_stmt_count.
	defer stmt.Close()

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
