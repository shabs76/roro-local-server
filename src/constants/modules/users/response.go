package users

type UserData struct {
	UserID       string `json:"userId" binding:"required"`
	FName        string `json:"fname" binding:"required"`
	LName        string `json:"lname" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required"`
	Password     string `json:"password" binding:"required,min=8"`
	RoleID       string `json:"roleId" binding:"required"`
	RoleName     string `json:"roleName" binding:"required"`
	RoleNumber   int    `json:"roleNumber" binding:"required"`
	Status       string `json:"status" binding:"required"`
	CreationDate string `json:"creationDate" binding:"required"`
}

type UserRole struct {
	RoleID     string `json:"roleId" binding:"required"`
	RoleName   string `json:"roleName" binding:"required"`
	RoleNumber int    `json:"roleNumber" binding:"required"`
	RoleDate   string `json:"roleDate" binding:"required"`
}

type UserLoginResponse struct {
	LoginID  string `json:"login_id" binding:"required"`
	LoginKey string `json:"login_key" binding:"required"`
	State    string `json:"state" binding:"required"`
	User     struct {
		UserID       string `json:"userId" binding:"required"`
		FName        string `json:"fname" binding:"required"`
		LName        string `json:"lname" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		Phone        string `json:"phone" binding:"required"`
		RoleID       string `json:"roleId" binding:"required"`
		RoleName     string `json:"roleName" binding:"required"`
		RoleNumber   int    `json:"roleNumber" binding:"required"`
		Status       string `json:"status" binding:"required"`
		CreationDate string `json:"creationDate" binding:"required"`
	} `json:"user" binding:"required"`
}

type UserLoginData struct {
	LogId      string `json:"logId" binding:"required"`
	LoginID    string `json:"loginId" binding:"required"`
	LoginKey   string `json:"loginKey" binding:"required"`
	UserID     string `json:"userId" binding:"required"`
	Status     string `json:"status" binding:"required"`
	ExpireDate string `json:"expireDate" binding:"required"`
	LoginDate  string `json:"loginDate" binding:"required"`
}
