package usersdataservices

import (
	"fmt"
	"strings"

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

// InsertUserData upserts users. The remote roles list holds only roles numbered above
// 100, but users of the other roles come in the users list too, so each user's role
// is added from the user record when it is not here yet (users.role is a foreign key
// to roles). A user that cannot be saved is reported and skipped; the others are still
// saved.
func InsertUserData(list []users.UserData) (st *constants.AnswerState) {
	roleQr := "INSERT INTO `roles`(`role_id`, `role_name`, `role_number`, `role_date`) VALUES (?,?,?,NOW()) ON DUPLICATE KEY UPDATE `role_id`=`role_id`"
	qr := "INSERT INTO `users`(`user_id`, `fname`, `lname`, `email`, `phone`, `password`, `role`, `status`, `creation_date`) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `fname`=VALUES(`fname`), `lname`=VALUES(`lname`), `email`=VALUES(`email`), `phone`=VALUES(`phone`), `password`=VALUES(`password`), `role`=VALUES(`role`), `status`=VALUES(`status`), `creation_date`=VALUES(`creation_date`)"

	failed := []string{}
	for _, user := range list {
		if user.RoleID != "" {
			gendb.SaveGeneral(roleQr, []any{user.RoleID, user.RoleName, user.RoleNumber})
		}
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
		if stx := gendb.SaveGeneral(qr, vals); stx.State != constants.SuccessState {
			failed = append(failed, fmt.Sprintf("%s (role %s)", user.Email, user.RoleName))
		}
	}

	if len(failed) > 0 && len(failed) == len(list) {
		return &constants.AnswerState{State: constants.ErrorState, Data: fmt.Sprintf("none of the %d users could be saved: %s", len(list), strings.Join(failed, ", ")), Adv: "none"}
	}
	if len(failed) > 0 {
		return &constants.AnswerState{State: constants.SuccessState, Data: fmt.Sprintf("%d of %d users synced; not saved: %s", len(list)-len(failed), len(list), strings.Join(failed, ", ")), Adv: "partial"}
	}
	return &constants.AnswerState{State: constants.SuccessState, Data: fmt.Sprintf("%d users were successfully synced", len(list)), Adv: "nothing"}
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
