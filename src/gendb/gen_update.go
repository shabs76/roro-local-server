package gendb

import (
	"log/slog"

	"github.com/shabs76/roro-local-server/constants"
)

func UpdateGeneral(sq string, vals []any) *constants.AnswerState {
	db, er := InitDb()

	if er != nil {
		slog.Error(er.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  er.Error(),
			Adv:   "none",
		}
	}

	res, err := db.Exec(sq, vals...)
	if err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to execute this request. Please try again",
			Adv:   "none",
		}
	}

	num, erN := res.RowsAffected()
	if erN != nil {
		slog.Error(erN.Error())
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to verify changes. Please try again",
			Adv:   "none",
		}
	}

	if num <= 0 {
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "process was successful but no change was done",
			Adv:   "okay",
		}
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "Update was successfully done",
		Adv:   "none",
	}
}
