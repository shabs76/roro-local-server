package usersdataservices

import (
	"fmt"

	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/users"
	"github.com/shabs76/roro-local-server/gendb"
	"github.com/shabs76/roro-local-server/specials"
)

func InsertUserRole(roles []users.UserRole) (st *constants.AnswerState) {
	qr := "INSERT INTO `roles`(`role_id`, `role_name`, `role_number`, `role_date`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `role_name`=VALUES(`role_name`), `role_number`=VALUES(`role_number`), `role_date`=VALUES(`role_date`)"
	for _, role := range roles {
		vals := []any{
			role.RoleID,
			role.RoleName,
			role.RoleNumber,
			role.RoleDate,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: fmt.Sprintf("%d roles were successfully synced", len(roles)), Adv: "nothing"}
}

func InsertUserData(users []users.UserData) (st *constants.AnswerState) {
	qr := "INSERT INTO `users`(`user_id`, `fname`, `lname`, `email`, `phone`, `password`, `role`, `status`, `creation_date`) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `fname`=VALUES(`fname`), `lname`=VALUES(`lname`), `email`=VALUES(`email`), `phone`=VALUES(`phone`), `password`=VALUES(`password`), `role`=VALUES(`role`), `status`=VALUES(`status`), `creation_date`=VALUES(`creation_date`)"

	for _, user := range users {
		vals := []any{
			user.UserID,
			user.FName,
			user.LName,
			user.Email,
			user.Phone,
			user.Password,
			user.RoleID,
			user.Status,
			user.CreationDate,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: fmt.Sprintf("%d users were successfully synced", len(users)), Adv: "nothing"}
}

func UserlonginsData(logs []users.UserLoginResponse) (st *constants.AnswerState) {
	qr := "INSERT INTO `logins`(`log_id`, `login_id`, `login_key`, `user_id`, `status`, `expire_date`, `login_date`) VALUES (?,?,?,?,?,DATE_ADD(NOW(), INTERVAL 47 HOUR),NOW()) ON DUPLICATE KEY UPDATE `login_key`=VALUES(`login_key`), `status`=VALUES(`status`), `expire_date`=VALUES(`expire_date`), `login_date`=VALUES(`login_date`)"
	id := specials.RandomString(36, "LOG")
	for _, log := range logs {
		vals := []any{
			id,
			log.LoginID,
			log.LoginKey,
			log.User.UserID,
			constants.StatusTypes.Active,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: fmt.Sprintf("%d user logins were successfully synced", len(logs)), Adv: "nothing"}
}
