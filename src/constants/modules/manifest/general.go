package manifest

type inspectionMarkStatuses struct {
	Missing       string
	Damaged       string
	SeenOrOkay    string
	NotApplicable string
}

var InspectionMarkStatus = inspectionMarkStatuses{
	Missing:       "missing",
	Damaged:       "damaged",
	SeenOrOkay:    "seen/okay",
	NotApplicable: "not-applicable",
}

type driverStatus struct {
	Active  string
	Blocked string
	Deleted string
}

var DriverStatus = driverStatus{
	Active:  "active",
	Blocked: "blocked",
	Deleted: "deleted",
}

type packageStatus struct {
	Active  string
	Blocked string
	Deleted string
}

var PackageStatus = packageStatus{
	Active:  "active",
	Blocked: "blocked",
	Deleted: "deleted",
}

type inspectionStatus struct {
	No  string
	Yes string
}

var InspectionStatus = inspectionStatus{
	No:  "no",
	Yes: "yes",
}

type genStatus struct {
	Active  string
	Blocked string
	Deleted string
}

var GenStatus = genStatus{
	Active:  "active",
	Blocked: "blocked",
	Deleted: "deleted",
}

type remarkStatus struct {
	Damaged string
	Info    string
}

var RemarkStatus = remarkStatus{
	Damaged: "damage",
	Info:    "info",
}
