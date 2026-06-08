package manifest

type InspectionCheck struct {
	CheckId     string `json:"checkId" binding:"required"`
	CheckStatus string `json:"checkValue" binding:"required"`
	FaultImage  string `json:"faultImage"`
}

type RemarkSaveRequest struct {
	Remark      string `json:"remark" binding:"required"`
	RemarkType  string `json:"remarkType" binding:"required" oneof:"info,damage"`
	RemarkImage string `json:"remarkImage" binding:"required"`
}

type VehicleExtraMediaRequest struct {
	MediaLink string `json:"mediaLink" binding:"required"`
	MediaType string `json:"mediaType" binding:"required" oneof:"image,video,pdf"`
	Remark    string `json:"remark" binding:"required"`
}

type OnBoardPackageMediaRequest struct {
	MediaLink string `json:"mediaLink" binding:"required"`
	MediaType string `json:"mediaType" binding:"required" oneof:"image,video,pdf"`
}

type OnBoardPackageRequest struct {
	Title  string                       `json:"title" binding:"required"`
	Remark string                       `json:"remark" binding:"required"`
	Media  []OnBoardPackageMediaRequest `json:"packageMedia" binding:"required"`
}

type InspectionChecksRequest struct {
	VehicleId      string                     `json:"vehicleId" binding:"required"`
	VehicleImage   string                     `json:"vehicleImage" binding:"required"`
	ManifestId     string                     `json:"manifestId" binding:"required"`
	MakerId        string                     `json:"makerId" binding:"required"`
	BodyId         string                     `json:"bodyId" binding:"required"`
	ModelName      string                     `json:"modelName" binding:"required"`
	InspectionTime string                     `json:"inspectionTime" binding:"required"`
	DeckNumber     string                     `json:"deckNumber" binding:"required"`
	NumberOfKeys   int                        `json:"numberOfKeys" binding:"min=0"`
	KeyType        string                     `json:"keyType" binding:"required"`
	Checks         []InspectionCheck          `json:"checks" binding:"required"`
	Remarks        []RemarkSaveRequest        `json:"remarks" binding:"required"`
	Media          []VehicleExtraMediaRequest `json:"media" binding:"required"`
	Packages       []OnBoardPackageRequest    `json:"onBoardPackage" binding:"required"`
}

type AddVehicleRequest struct {
	ManifestId   string  `json:"manifestId" binding:"required"`
	ChasisNumber string  `json:"chasisNumber" binding:"required"`
	VehicleModel string  `json:"vehicleModel" binding:"required"`
	Description  string  `json:"description" binding:"required"`
	Weight       float64 `json:"weight" binding:"required"`
	IsOverLand   string  `json:"isOverLand" binding:"required" oneof:"yes,no"`
	BLNumber     string  `json:"blNumber" binding:"required"`
}

// Independent package
type AddPackageRequest struct {
	ManifestId    string `json:"manifestId" binding:"required"`
	BLNumber      string `json:"blNumber" binding:"required"`
	PackageNumber string `json:"packageNumber" binding:"required"`
	Description   string `json:"description" binding:"required"`
}

type PackageInspectionSaveRequest struct {
	PackageId          string                     `json:"packageId" binding:"required"`
	PackageImage       string                     `json:"packageImage" binding:"required"`
	TypeId             string                     `json:"typeId" binding:"required"`
	InspectionTime     string                     `json:"inspectionTime" binding:"required"`
	InspectionStatusId string                     `json:"inspectionStatusId" binding:"required"`
	Media              []PackageExtraMediaRequest `json:"media" binding:"required"`
}

type PackageExtraMediaRequest struct {
	MediaLink string `json:"mediaLink" binding:"required"`
	MediaType string `json:"mediaType" binding:"required" oneof:"image,video,pdf"`
	Remark    string `json:"remark" binding:"required"`
}

type VehicleRemarksOnlyRequest struct {
	VehicleId string              `json:"vehicleId" binding:"required"`
	Remarks   []RemarkSaveRequest `json:"remarks" binding:"required,min=1"`
}
