package manifest

type VehicleBodyTypes struct {
	BodyId       string `json:"bodyId" binding:"required"`
	BodyName     string `json:"bodyName" binding:"required"`
	CreationDate string `json:"creationDate" binding:"required"`
}

type VehicleMakers struct {
	MakerId      string `json:"makerId" binding:"required"`
	MakerName    string `json:"makerName" binding:"required"`
	CreationDate string `json:"creationDate" binding:"required"`
}

type InspectionCheckListDetails struct {
	CheckId        string `json:"checkId" binding:"required"`
	CheckName      string `json:"checkName" binding:"required"`
	UpdatedDate    string `json:"updatedDate" binding:"required"`
	RegisteredDate string `json:"registeredDate" binding:"required"`
}

type VehicleModelDetails struct {
	ModelId      string `json:"modelId" binding:"required"`
	ModelName    string `json:"modelName" binding:"required"`
	Status       string `json:"status" binding:"required"`
	CreationDate string `json:"creationDate" binding:"required"`
}

type PackageTypeDetails struct {
	TypeId      string `json:"typeId" binding:"required"`
	TypeName    string `json:"typeName" binding:"required"`
	Status      string `json:"status" binding:"required"`
	CreatedTime string `json:"createdTime" binding:"required"`
}

type PackageInspectionStatusDetails struct {
	StatusId   string `json:"statusId" binding:"required"`
	StatusName string `json:"statusName" binding:"required"`
	StatusDesc string `json:"statusDesc" binding:"required"`
	StatusNum  int    `json:"statusNum" binding:"required"`
}

type ClientInfo struct {
	ClientID       string `json:"clientId" validate:"required" binding:"required"`
	ClientName     string `json:"clientName" validate:"required" binding:"required"`
	Logo           string `json:"logo" validate:"required" binding:"required"`
	Cover          string `json:"cover" validate:"required" binding:"required"`
	PrincipalName  string `json:"principalName" validate:"required" binding:"required"`
	ClientLocation string `json:"clientLocation" validate:"required" binding:"required"`
	Status         string `json:"status" validate:"required" binding:"required"`
	CreationDate   string `json:"creationDate" validate:"required" binding:"required"`
}

// MANIFEST RELATED RESPONSES
type DeckStowagePlanDetails struct {
	DeckId     string `json:"deckId" binding:"required"`
	DeckName   string `json:"deckName" binding:"required"`
	ManifestId string `json:"manifestId" binding:"required"`
	Units      int    `json:"units" binding:"required"`
}

type VehicleData struct {
	VehicleId        string  `json:"vehicleId" binding:"required"`
	ManifestId       string  `json:"manifestId" binding:"required"`
	ChasisNumber     string  `json:"chasisNumber" binding:"required"`
	VehicleModel     string  `json:"vehicleModel" binding:"required"`
	Description      string  `json:"description" binding:"required"`
	Weight           float32 `json:"weight" binding:"required"`
	BLNumber         string  `json:"blNumber" binding:"required"`
	CreationDate     string  `json:"creationDate" binding:"required"`
	InspectionStatus string  `json:"inspectionStatus" binding:"required"`
	TalliedStatus    string  `json:"talliedStatus" binding:"required"`
	DischargeStatus  string  `json:"dischargeStatus" binding:"required"`
	InspectionTime   string  `json:"inspectionTime" binding:"required"`
	TalliedTime      string  `json:"talliedTime" binding:"required"`
	DischargeTime    string  `json:"dischargeTime" binding:"required"`
	OverLandStatus   string  `json:"overLandStatus" binding:"required"`
	IsAddedLater     string  `json:"isAddedLater"`
	IsPublished      string  `json:"isPublished"`
}

type VehiclesDetailsToShow struct {
	VehicleId        string  `json:"vehicleId" binding:"required"`
	ManifestId       string  `json:"manifestId" binding:"required"`
	ChasisNumber     string  `json:"chasisNumber" binding:"required"`
	VehicleModel     string  `json:"vehicleModel" binding:"required"`
	Description      string  `json:"description" binding:"required"`
	Weight           float32 `json:"weight" binding:"required"`
	CreationDate     string  `json:"creationDate" binding:"required"`
	InspectionStatus string  `json:"inspectionStatus" binding:"required"`
	TalliedStatus    string  `json:"talliedStatus" binding:"required"`
	DischargeStatus  string  `json:"dischargeStatus" binding:"required"`
	InspectionTime   string  `json:"inspectionTime" binding:"required"`
	TalliedTime      string  `json:"talliedTime" binding:"required"`
	DischargeTime    string  `json:"dischargeTime" binding:"required"`
	Maker            string  `json:"maker" binding:"required"`
	BodyType         string  `json:"bodyType" binding:"required"`
	Image            string  `json:"image"`
	DeckNumber       string  `json:"deckNumber" binding:"required"`
	IsDamaged        string  `json:"isDamaged" binding:"required"`
	IsOverLand       string  `json:"isOverLand" binding:"required"`
	BLNumber         string  `json:"blNumber" binding:"required"`
	NumberOfKeys     int     `json:"numberOfKeys" binding:"required"`
}

type ManifestData struct {
	ManifestId   string `json:"manifestId" binding:"required"`
	ManifestName string `json:"manifestName" binding:"required"`
	ClientId     string `json:"clientId" binding:"required"`
	ClientName   string `json:"clientName" binding:"required"`
	VesselName   string `json:"vesselName" binding:"required"`
	VoyageNo     string `json:"voyageNo" binding:"required"`
	BerthNo      string `json:"berthNo" binding:"required"`
	ArrivalDate  string `json:"arrivalDate" binding:"required"`
	ReceivedDate string `json:"receivedDate" binding:"required"`
	UploadedDate string `json:"uploadedDate" binding:"required"`
}

type ManifestToshow struct {
	ManifestId        string `json:"manifestId" binding:"required"`
	ManifestName      string `json:"manifestName" binding:"required"`
	ClientName        string `json:"clientName" binding:"required"`
	VesselName        string `json:"vesselName" binding:"required"`
	VoyageNo          string `json:"voyageNo" binding:"required"`
	BerthNo           string `json:"berthNo" binding:"required"`
	ArrivalDate       string `json:"arrivalDate" binding:"required"`
	ReceivedDate      string `json:"receivedDate" binding:"required"`
	UploadedDate      string `json:"uploadedDate" binding:"required"`
	VehiclesNumber    int    `json:"vehiclesNumber" binding:"required"`
	Inspected         int    `json:"vehiclesInspected" binding:"required"`
	VehiclesDamaged   int    `json:"vehiclesDamaged" binding:"required"`
	VehicleDischarged int    `json:"vehiclesDischarged" binding:"required"`
	VehicleRemarks    int    `json:"vehicleRemarks" binding:"required"`
	TotalPackages     int    `json:"totalPackages" binding:"required"`
	InspectedPackages int    `json:"inspectedPackages" binding:"required"`
}

type PackageManifestInfo struct {
	PackageId     string `json:"packageId" binding:"required"`
	PackageNumber string `json:"packageNumber" binding:"required"`
	ManifestId    string `json:"manifestId" binding:"required"`
	BLNumber      string `json:"blNumber" binding:"required"`
	Description   string `json:"description" binding:"required"`
	IsInspected   string `json:"isInspected" binding:"required"`
	IsAddedLater  string `json:"isAddedLater" binding:"required"`
	CreationTime  string `json:"creationTime" binding:"required"`
	IsPublished   string `json:"isPublished"`
}

type PackageInspectionDetails struct {
	ManifestId            string                `json:"manifestId" binding:"required"`
	PackageId             string                `json:"packageId" binding:"required"`
	BLNumber              string                `json:"blNumber" binding:"required"`
	PackageNumber         string                `json:"packageNumber" binding:"required"`
	Description           string                `json:"description" binding:"required"`
	IsInspected           string                `json:"isInspected" binding:"required"`
	InspectionId          string                `json:"inspectionId" binding:"required"`
	TypeId                string                `json:"typeId" binding:"required"`
	TypeName              string                `json:"typeName" binding:"required"`
	Picture               string                `json:"picture" binding:"required"`
	InspectionStatusId    string                `json:"inspectionStatusId" binding:"required"`
	InspectionStatus      string                `json:"inspectionStatus" binding:"required"`
	InspectionDescription string                `json:"inspectionDescription" binding:"required"`
	InspectionNumber      int                   `json:"inspectionNumber" binding:"required"`
	InspectionTime        string                `json:"inspectionTime" binding:"required"`
	CreationTime          string                `json:"creationTime" binding:"required"`
	UserId                string                `json:"userId" binding:"required"`
	UserFname             string                `json:"userFirstName" binding:"required"`
	UserLname             string                `json:"userLastName" binding:"required"`
	UserPhone             string                `json:"userPhone" binding:"required"`
	IsAddedLater          string                `json:"isAddedLater"`
	Media                 []PackageMediaDetails `json:"media" binding:"required"`
}

type PackageMediaDetails struct {
	MediaId   string `json:"mediaId" binding:"required"`
	PackageId string `json:"packageId" binding:"required"`
	MediaType string `json:"mediaType" binding:"required"`
	MediaLink string `json:"mediaLink" binding:"required"`
	Remark    string `json:"remark" binding:"required"`
	Status    string `json:"status" binding:"required"`
}
type VehiclesDetailsAndInspection struct {
	VehicleId           string  `json:"vehicleId" binding:"required"`
	ManifestId          string  `json:"manifestId" binding:"required"`
	ChasisNumber        string  `json:"chasisNumber" binding:"required"`
	VehicleModel        string  `json:"vehicleModel" binding:"required"`
	Description         string  `json:"description" binding:"required"`
	Weight              float32 `json:"weight" binding:"required"`
	BLNumber            string  `json:"blNumber" binding:"required"`
	CreationDate        string  `json:"creationDate" binding:"required"`
	InspectionStatus    string  `json:"inspectionStatus" binding:"required"`
	TalliedStatus       string  `json:"talliedStatus" binding:"required"`
	DischargeStatus     string  `json:"dischargeStatus" binding:"required"`
	OverLandStatus      string  `json:"overLandStatus" binding:"required"`
	InspectionTime      string  `json:"inspectionTime" binding:"required"`
	TalliedTime         string  `json:"talliedTime" binding:"required"`
	DischargeTime       string  `json:"dischargeTime" binding:"required"`
	InspectionMark      string  `json:"inspectionMark" binding:"required"`
	InspectionCheckId   string  `json:"inspectionCheckId" binding:"required"`
	InspectionCheckName string  `json:"inspectionCheckName" binding:"required"`
}

type InspectionImageDetails struct {
	ImageId      string `json:"imageId" binding:"required"`
	ImageLink    string `json:"imageLink" bindihg:"required"`
	InspectionId string `json:"inspectionId" binding:"required"`
	CreationTime string `json:"creationTime" binding:"required"`
}

type InspectionRemarksDetails struct {
	RemarkId   string `json:"remarkId" binding:"required"`
	VehicleId  string `json:"vehicleId" binding:"required"`
	UserId     string `json:"userId" binding:"required"`
	ImageLink  string `json:"remarkImage" binding:"required"`
	Remark     string `json:"remark" binding:"required"`
	RemarkType string `json:"remarkType" binding:"required"`
	RemarkTime string `json:"remarkTime" binding:"required"`
}

type RemarkImageDetails struct {
	ImageId      string `json:"imageId" binding:"required"`
	RemarkId     string `json:"remarkId" binding:"required"`
	ImageLink    string `json:"imageLink" binding:"required"`
	UploadedDate string `json:"uploadedDate" binding:"required"`
}

type InspectionDetails struct {
	InspectionId string `json:"inspectionId" binding:"required"`
	VehicleId    string `json:"vehicleId" binding:"required"`
	CheckId      string `json:"checkId" binding:"required"`
	CheckName    string `json:"checkName" binding:"require"`
	Status       string `json:"status" binding:"required"`
	Checktime    string `json:"checkTime" binding:"required"`
}

type InspectionDetailsAndImage struct {
	InspectionId string `json:"inspectionId" binding:"required"`
	VehicleId    string `json:"vehicleId" binding:"required"`
	CheckId      string `json:"checkId" binding:"required"`
	CheckName    string `json:"checkName" binding:"require"`
	UserId       string `json:"userId" binding:"required"`
	Status       string `json:"status" binding:"required"`
	Checktime    string `json:"checkTime" binding:"required"`
	Image        string `json:"image" binding:"required"`
}

type TallyDetails struct {
	TallyId      string `json:"tallyId" binding:"required"`
	VehicleId    string `json:"vehicleId" binding:"required"`
	ManifestId   string `json:"manifestId" binding:"required"`
	MakerId      string `json:"makerId" binding:"required"`
	BodyId       string `json:"bodyId" binding:"required"`
	VehicleImage string `json:"vehicleImage" binding:"required"`
	DeckNumber   string `json:"deckNumber" binding:"required"`
	NumberOfKeys int    `json:"numberOfKeys" binding:"required"`
	KeyType      string `json:"keyType" binding:"required"`
	TalliedTime  string `json:"talliedTime" binding:"required"`
}

type TallyMoreDetails struct {
	TallyId      string `json:"tallyId" binding:"required"`
	VehicleId    string `json:"vehicleId" binding:"required"`
	ManifestId   string `json:"manifestId" binding:"required"`
	MakerId      string `json:"makerId" binding:"required"`
	UserId       string `json:"userId" binding:"required"`
	BodyId       string `json:"bodyId" binding:"required"`
	TalliedTime  string `json:"talliedTime" binding:"required"`
	MakerName    string `json:"makerName" binding:"required"`
	BodyName     string `json:"bodyName" binding:"required"`
	VehicleImage string `json:"vehicleImage" binding:"required"`
	DeckNumber   string `json:"deckNumber" binding:"required"`
	NumberOfKeys int    `json:"numberOfKeys" binding:"required"`
	KeyType      string `json:"keyType" binding:"required"`
}

type VehiclesDetailsAndTally struct {
	VehicleId        string  `json:"vehicleId" binding:"required"`
	ManifestId       string  `json:"manifestId" binding:"required"`
	UserId           string  `json:"userId" binding:"required"`
	ChasisNumber     string  `json:"chasisNumber" binding:"required"`
	VehicleModel     string  `json:"vehicleModel" binding:"required"`
	Description      string  `json:"description" binding:"required"`
	Weight           float32 `json:"weight" binding:"required"`
	BLNumber         string  `json:"blNumber" binding:"required"`
	CreationDate     string  `json:"creationDate" binding:"required"`
	InspectionStatus string  `json:"inspectionStatus" binding:"required"`
	TalliedStatus    string  `json:"talliedStatus" binding:"required"`
	DischargeStatus  string  `json:"dischargeStatus" binding:"required"`
	InspectionTime   string  `json:"inspectionTime" binding:"required"`
	TalliedTime      string  `json:"talliedTime" binding:"required"`
	DischargeTime    string  `json:"dischargeTime" binding:"required"`
	OverLandStatus   string  `json:"overLandStatus" binding:"required"`
	IsAddedLater     string  `json:"isAddedLater" binding:"required"`
	// tally data
	TallyId      string `json:"tallyId" binding:"required"`
	MakerId      string `json:"makerId" binding:"required"`
	BodyId       string `json:"bodyId" binding:"required"`
	MakerName    string `json:"makerName" binding:"required"`
	BodyName     string `json:"bodyName" binding:"required"`
	VehicleImage string `json:"vehicleImage" binding:"required"`
	DeckNumber   string `json:"deckNumber" binding:"required"`
	NumberOfKeys int    `json:"numberOfKeys" binding:"required"`
	KeyType      string `json:"keyType" binding:"required"`
}

type OnBoardPackageMediaResp struct {
	MediaId   string `json:"mediaId" binding:"required"`
	MediaType string `json:"mediaType" binding:"required"`
	MediaLink string `json:"mediaLink" binding:"required"`
	PackageId string `json:"packageId" binding:"required"`
}

type OnboardPackageResp struct {
	PackageId string                    `json:"packageId" binding:"required"`
	Title     string                    `json:"title" binding:"required"`
	Remark    string                    `json:"remark" binding:"required"`
	VehicleId string                    `json:"vehicleId" binding:"required"`
	Media     []OnBoardPackageMediaResp `json:"media" binding:"required"`
}

type VehicleMediaResp struct {
	MediaId   string `json:"mediaId" binding:"required"`
	MediaLink string `json:"mediaLink" binding:"required"`
	MediaType string `json:"mediaType" binding:"required"`
	Remark    string `json:"remark" binding:"required"`
	VehicleId string `json:"vehicleId" binding:"required"`
	Status    string `json:"status" binding:"required"`
}
type VehicleDischargeDetails struct {
	DischargeId    string `json:"dischargeId" binding:"required"`
	VehicleId      string `json:"vehicleId" binding:"required"`
	DischargeImage string `json:"dischargeImage" binding:"required"`
	UserId         string `json:"userId" binding:"required"`
	UserFname      string `json:"userFirstName" binding:"required"`
	UserLname      string `json:"userLastName" binding:"required"`
	UserPhone      string `json:"userPhone" binding:"required"`
	DriverId       string `json:"driverId" binding:"required"`
	DriverFname    string `json:"driverFirstName" binding:"required"`
	DriverLname    string `json:"driverLastName" binding:"required"`
	DriverPhone    string `json:"driverPhone" binding:"required"`
	DischargeTime  string `json:"dischargeTime" binding:"required"`
}

type MakerSummary struct {
	MakerName string `json:"makerName"`
	Count     int    `json:"count"`
}

type BodySummary struct {
	BodyName string `json:"bodyName"`
	Count    int    `json:"count"`
}

// Combined summary for both maker and body type
type VehicleSummary struct {
	Makers     []MakerSummary `json:"makers"`
	BodyTypes  []BodySummary  `json:"bodyTypes"`
	TotalCount int            `json:"totalCount"`
	OverLand   int            `json:"overLand"`
	Inspected  int            `json:"inspected"`
	Discharged int            `json:"discharged"`
}
