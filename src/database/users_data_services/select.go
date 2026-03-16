package usersdataservices

import (
	"log"

	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/users"
	"github.com/shabs76/roro-local-server/gendb"
)

func SelectUserDetailsWithRolesPass(subQuery string, vals []any) (*constants.AnswerState, []users.UserData) {
	qr := "SELECT `user_id`, `fname`, `lname`, `email`, `phone`, `password`, `role`, `role_name`, `role_number`, `status`, `creation_date` FROM `users` INNER JOIN roles ON roles.role_id = users.role WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var usersList []users.UserData

	for res.Next() {
		var user users.UserData
		ers := res.Scan(&user.UserID, &user.FName, &user.LName, &user.Email, &user.Phone, &user.Password, &user.RoleID, &user.RoleName, &user.RoleNumber, &user.Status, &user.CreationDate)
		if ers != nil {
			log.Println(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind user details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		usersList = append(usersList, user)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, usersList
}

func SelectUserRoles(subQuery string, vals []any) (*constants.AnswerState, []users.UserRole) {
	qr := "SELECT `role_id`, `role_name`, `role_number`, `role_date` FROM `roles` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rolesList []users.UserRole

	for res.Next() {
		var role users.UserRole
		ers := res.Scan(&role.RoleID, &role.RoleName, &role.RoleNumber, &role.RoleDate)
		if ers != nil {
			log.Println(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind user role details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		rolesList = append(rolesList, role)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rolesList
}

func SelectUserLogins(subQuery string, vals []any) (*constants.AnswerState, []users.UserLoginData) {
	qr := "SELECT `log_id`, `login_id`, `login_key`, `user_id`, `status`, `expire_date`, `login_date` FROM `logins` WHERE " + subQuery
	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var loginsList []users.UserLoginData

	for res.Next() {
		var logins users.UserLoginData
		ers := res.Scan(&logins.LogId, &logins.LoginID, &logins.LoginKey, &logins.UserID, &logins.Status, &logins.ExpireDate, &logins.LoginDate)
		if ers != nil {
			log.Println(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind user login details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		loginsList = append(loginsList, logins)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, loginsList
}
