package controlers

import (
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/page"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/linestyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
	usersdataservices "github.com/shabs76/roro-local-server/database/users_data_services"
	"github.com/shabs76/roro-local-server/specials"
)

// Define color helper functions before they are used
func getWallOrange() *props.Color {
	return &props.Color{
		Red:   255,
		Green: 109,
		Blue:  1,
	}
}

func getGrayColor() *props.Color {
	return &props.Color{
		Red:   170,
		Green: 170,
		Blue:  170,
	}
}

func getWallsGreen() *props.Color {
	return &props.Color{
		Red:   40,
		Green: 167,
		Blue:  69,
	}
}

func getRedColor() *props.Color {
	return &props.Color{
		Red:   150,
		Green: 10,
		Blue:  10,
	}
}

func getBlueColor() *props.Color {
	return &props.Color{
		Red:   10,
		Green: 10,
		Blue:  150,
	}
}

func GenerateDamagedVahiclePdf(c *gin.Context) {

	mId := c.Param("manifestId")
	// `inspection_time`, `tallied_time`
	subQr := " `manifest_id` = ? AND (vehicles_inspection.status = ? OR vehicles_inspection.status = ?)"
	vals := []any{mId, manifest.InspectionMarkStatus.Damaged, manifest.InspectionMarkStatus.Missing}

	st, vehicles := manifestdataservices.SelectVehiclesAndInspectionDetails(subQr+" GROUP BY manifest_vehicles.vehicle_id   ORDER BY inspection_time ASC, tallied_time ASC", vals)
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "failed to obtain vehicles list"})
		return
	} else if len(vehicles) <= 0 {
		c.JSON(http.StatusNoContent, gin.H{"state": constants.SuccessState, "data": []any{}})
		return
	}

	var inspDets = []VehicleInspectionDetailsRes{}

	// loop to check damage status on the vehicle
	for _, vehicle := range vehicles {
		// Get vehicle inspection details
		st, vehicleDets := manifestdataservices.SelectVehicleAndTallyDetails(" manifest_vehicles.vehicle_id = ? AND `inspection_status` = ?", []any{vehicle.VehicleId, "yes"})
		if st.State != constants.SuccessState {
			slog.Error(st.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle details due to " + st.Data})
			return
		} else if len(vehicleDets) <= 0 {
			continue
		}

		// get inspection detail list
		sti, insps := manifestdataservices.SelectInspectionDetails(" `vehicle_id` = ? ", []any{vehicle.VehicleId})
		if sti.State != constants.SuccessState {
			slog.Error(sti.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle inspection details due to " + sti.Data})
			return
		}
		var inspImg = []manifest.InspectionDetailsAndImage{}
		for i := range insps {
			img := ""
			stm, imgs := manifestdataservices.SelectInspectionImages(" `inspection_id` = ?", []any{insps[i].InspectionId})
			if stm.State != constants.SuccessState {
				slog.Error(stm.Data)
				c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch inspection images due to " + stm.Data})
				return
			} else if len(imgs) > 0 {
				img = imgs[0].ImageLink
			}
			inspImg = append(inspImg, manifest.InspectionDetailsAndImage{
				InspectionId: insps[i].InspectionId,
				VehicleId:    insps[i].VehicleId,
				CheckId:      insps[i].CheckId,
				CheckName:    insps[i].CheckName,
				UserId:       vehicleDets[0].UserId,
				Status:       insps[i].Status,
				Checktime:    insps[i].Checktime,
				Image:        img,
			})

		}
		// get user details
		stu, users := usersdataservices.SelectUserDetailsWithRolesPass(" `user_id` = ? ", []any{inspImg[0].UserId})
		if stu.State != constants.SuccessState {
			slog.Error(stu.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch user details due to " + stu.Data})
			return
		} else if len(users) <= 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch user details due to user not found"})
			return
		}

		// get inspection remarks
		str, rem := manifestdataservices.SelectInspectionRemarks(" `vehicle_id` = ? AND `remark_type` = ?", []any{vehicle.VehicleId, manifest.RemarkStatus.Damaged})
		if str.State != constants.SuccessState {
			slog.Error(str.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle inspection remarks details due to " + str.Data})
			return
		}

		discharge := []manifest.VehicleDischargeDetails{}
		// discharge image
		stim, dis := manifestdataservices.SelectDischargeDetails(" `vehicle_id` = ? ", []any{vehicle.VehicleId})
		if stim.State != constants.SuccessState {
			slog.Error(stim.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle discharge details details due to " + stim.Data})
			return
		} else if len(dis) > 1 {
			discharge = dis
		}

		inspDets = append(inspDets, VehicleInspectionDetailsRes{
			VehicleDets:   vehicleDets[0],
			Inspections:   inspImg,
			Remarks:       rem,
			DischargeInfo: discharge,
			UserDetails:   users[0],
		})
	}

	// 2. Check damage vehicles based on remarks only
	strem, remarks := manifestdataservices.SelectInspectionRemarksForGivenManifest(" `manifest_id` = ? AND `remark_type` = ?  GROUP BY manifest_vehicles.vehicle_id", []any{mId, manifest.RemarkStatus.Damaged})
	if strem.State != constants.SuccessState {
		slog.Error(strem.Data)
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": constants.ErrorState,
			"data":  "Failed to fetch vehicle inspection remarksx details due to " + strem.Data,
		})
		return
	}

	// 3. filter vehicles which have already been found to have faults
	nofoundRemarks := make([]manifest.InspectionRemarksDetails, 0, len(remarks))
	for _, remark := range remarks {
		for _, insp := range inspDets {
			// check if the remark is already in the inspDets
			if insp.VehicleDets.VehicleId == remark.VehicleId {
				nofoundRemarks = append(nofoundRemarks, remark)
			}

		}
	}

	// 4. Build inspDets and add to main list based on the remarks results
	for _, remark := range nofoundRemarks {
		// Get vehicle inspection details
		st, vehicleDets := manifestdataservices.SelectVehicleAndTallyDetails(" manifest_vehicles.vehicle_id = ? AND `inspection_status` = ?", []any{remark.VehicleId, "yes"})
		if st.State != constants.SuccessState {
			slog.Error(st.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle details due to " + st.Data})
			return
		} else if len(vehicleDets) <= 0 {
			continue
		}

		// get inspection detail list
		sti, insps := manifestdataservices.SelectInspectionDetails(" `vehicle_id` = ? ", []any{remark.VehicleId})
		if sti.State != constants.SuccessState {
			slog.Error(sti.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle inspection details due to " + sti.Data})
			return
		}
		var inspImg = []manifest.InspectionDetailsAndImage{}
		for i := range insps {
			img := ""
			stm, imgs := manifestdataservices.SelectInspectionImages(" `inspection_id` = ?", []any{insps[i].InspectionId})
			if stm.State != constants.SuccessState {
				slog.Error(stm.Data)
				c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch inspection images due to " + stm.Data})
				return
			} else if len(imgs) > 0 {
				img = imgs[0].ImageLink
			}

			inspImg = append(inspImg, manifest.InspectionDetailsAndImage{
				InspectionId: insps[i].InspectionId,
				VehicleId:    insps[i].VehicleId,
				CheckId:      insps[i].CheckId,
				CheckName:    insps[i].CheckName,
				UserId:       vehicleDets[0].UserId,
				Status:       insps[i].Status,
				Checktime:    insps[i].Checktime,
				Image:        img,
			})
		}
		// get user details
		stu, users := usersdataservices.SelectUserDetailsWithRolesPass(" `user_id` = ? ", []any{inspImg[0].UserId})
		if stu.State != constants.SuccessState {
			slog.Error(stu.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch user details due to " + stu.Data})
			return
		} else if len(users) <= 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch user details due to user not found"})
			return
		}

		// get inspection remarks
		str, rem := manifestdataservices.SelectInspectionRemarks(" `vehicle_id` = ? AND `remark_type` = ? ", []any{remark.VehicleId, manifest.RemarkStatus.Damaged})
		if str.State != constants.SuccessState {
			slog.Error(str.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle inspection remarkx details due to " + str.Data})
			return
		}

		discharge := []manifest.VehicleDischargeDetails{}
		// discharge image
		stim, dis := manifestdataservices.SelectDischargeDetails(" `vehicle_id` = ? ", []any{remark.VehicleId})
		if stim.State != constants.SuccessState {
			slog.Error(stim.Data)
			c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle discharge details details due to " + stim.Data})
			return
		} else if len(dis) > 1 {
			discharge = dis
		}

		inspDets = append(inspDets, VehicleInspectionDetailsRes{
			VehicleDets:   vehicleDets[0],
			Inspections:   inspImg,
			Remarks:       rem,
			DischargeInfo: discharge,
			UserDetails:   users[0],
		})
	}

	// 5. Get manifest details
	st, manifests := manifestdataservices.SelectManifestInfo(" `manifest_id` = ? ", []any{mId})
	if st.State != constants.SuccessState {
		slog.Error(fmt.Sprintf("Error fetching manifest: %s", st.Data))
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": constants.ErrorState,
			"data":  "Failed to retrieve manifest information",
		})
		return
	}

	if len(manifests) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"state": constants.ErrorState,
			"data":  "Manifest does not exist",
		})
		return
	}

	manifest := manifests[0]

	m := generateFaultMaroto(inspDets, manifest)

	document, err := m.Generate()
	if err != nil {
		log.Fatal(err.Error())
	}

	pdfBytes := document.GetBytes()

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=fault_vehicle_report.pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func generateFaultMaroto(inspections []VehicleInspectionDetailsRes, manifest manifest.ManifestData) core.Maroto {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	err := m.RegisterHeader(getPageHeaderFault(manifest))
	if err != nil {
		log.Panic(err.Error())
	}

	err = m.RegisterFooter(getPageFooterFault()...)
	if err != nil {
		log.Panic(err.Error())
	}

	newPage := page.New()
	// loop every vehicle
	for i, insp := range inspections {
		if i > 0 {

			m.AddPages(newPage)
		}
		// heading
		m.AddRows(text.NewRow(16, "Vehicle Discharge Inspection and Technical Report (VDITR)", props.Text{
			Top:   5,
			Size:  16,
			Style: fontstyle.Bold,
			Align: align.Center,
		}),
			row.New(5),
		)
		m.AddRows(
			buildMainfestSectionFault(manifest)...,
		)

		m.AddRows(
			row.New(3),
		)

		// upper sections
		m.AddRows(
			buildUpperSections(insp, manifest)...,
		)
		// faults and remarks
		m.AddRows(
			buildFaultsAndRemarks(insp)...,
		)

		// section F Inspection Checklist
		m.AddPages(newPage)
		m.AddRows(
			buildInspectionCheklistTable(insp)...,
		)
	}

	return m
}

func buildInspectionCheklistTable(insp VehicleInspectionDetailsRes) []core.Row {
	rows := make([]core.Row, 0, len(insp.Inspections)+2) //
	// section F Inspection Checklist
	rows = append(rows,
		row.New(5),
		text.NewAutoRow("F. Inspection Checklist", props.Text{
			Size:  13,
			Style: fontstyle.Bold,
			Align: align.Left,
		}),
		row.New(3),
	)

	// header for the table
	rows = append(rows, row.New(10).Add(
		text.NewCol(1, "S/N", props.Text{
			Top:   2.5,
			Left:  2,
			Size:  10,
			Style: fontstyle.Bold,
			Align: align.Center,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(7, "Check Name", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(4, "Status", props.Text{
			Top:   2.5,
			Size:  10,
			Right: 3,
			Style: fontstyle.Bold,
			Align: align.Right,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))
	for i, check := range insp.Inspections {
		if i%2 != 0 {
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1), props.Text{
					Top:   2.5,
					Left:  2,
					Size:  10,
					Align: align.Center,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(7, check.CheckName, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(4, specials.CapitalizeWords(check.Status), props.Text{
					Top:   2.5,
					Size:  10,
					Right: 3,
					Align: align.Right,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
			))
		} else {
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1), props.Text{
					Top:   2.5,
					Left:  2,
					Size:  10,
					Align: align.Center,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(7, check.CheckName, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(4, specials.CapitalizeWords(check.Status), props.Text{
					Top:   2.5,
					Size:  10,
					Right: 3,
					Align: align.Right,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
			))
		}
	}

	rows = append(rows, row.New(5)) // spacer
	return rows

}

func buildFaultsAndRemarks(insp VehicleInspectionDetailsRes) []core.Row {
	rows := []core.Row{}
	// section D Faults and Remarks
	rows = append(rows,
		row.New(5),
		text.NewAutoRow("D. Faults and Remarks", props.Text{
			Size:  13,
			Style: fontstyle.Bold,
			Align: align.Left,
		}))

	// loop through check list and pick those with damage yes
	columns := []core.Col{}
	for _, check := range insp.Inspections {
		if check.Status == manifest.InspectionMarkStatus.Damaged || check.Status == manifest.InspectionMarkStatus.Missing {

			imageBytes, err := specials.GetImageBytesForPdf(check.Image)
			if err != nil {
				slog.Error(fmt.Sprintf("Error fetching/converting image for PDF: %s", err.Error()))
				continue
			}

			columns = append(columns, col.New(4).Add(
				image.NewFromBytes(imageBytes, extension.Png, props.Rect{
					Center:             false,
					Percent:            80,
					JustReferenceWidth: true,
				}),
				text.New(check.CheckName, props.Text{
					Top:   45,
					Size:  10,
					Color: getWallOrange(),
				}),
				text.New(check.Status, props.Text{
					Top:   55,
					Size:  10,
					Color: getRedColor(),
				}),
			))
		}
	}

	// loop through every remark and check
	for _, remark := range insp.Remarks {
		imageBytes, err := specials.GetImageBytesForPdf(remark.ImageLink)
		if err != nil {
			slog.Error(fmt.Sprintf("Error fetching/converting image for PDF: %s", err.Error()))
			continue
		}
		columns = append(columns, col.New(4).Add(

			image.NewFromBytes(imageBytes, extension.Png, props.Rect{
				Center:             false,
				Percent:            90,
				JustReferenceWidth: true,
			}),
			text.New(remark.Remark, props.Text{
				Top:   50,
				Size:  10,
				Color: getRedColor(),
			}),
		))
	}
	// loop to arrange them in a row. Each row should take three and when three is reached new row should be started
	// Calculate number of rows needed (ceiling of columns / 3)
	rowsNeeded := int(math.Ceil(float64(len(columns)) / 3.0))
	// Create rows with 3 columns each
	for i := range rowsNeeded {
		rows = append(rows, row.New(10)) // spacer
		start := i * 3
		end := min(start+3, len(columns))

		// Create a row with up to 3 columns
		currentRow := row.New(50).Add(columns[start:end]...)
		rows = append(rows, currentRow)
	}

	rows = append(rows, row.New(5)) // spacer
	return rows
}

func buildUpperSections(insp VehicleInspectionDetailsRes, manifest manifest.ManifestData) []core.Row {
	now := time.Now()
	layout := "2 Jan 2006"
	rows := []core.Row{}
	arrivalDate, err := time.Parse("2006-01-02", manifest.ArrivalDate)
	if err != nil {
		slog.Error(fmt.Sprintf("Error parsing arrival date: %s", err.Error()))
		arrivalDate = now // Fallback to current time if parsing fails
	}

	// section A General Particulars
	rows = append(rows, text.NewAutoRow("A. General Particulars", props.Text{

		Size:  13,
		Style: fontstyle.Bold,
		Align: align.Left,
	}),
		row.New(7).Add(
			col.New(6).Add(
				text.New("A1.Deck No.: "+strings.ToUpper(insp.VehicleDets.DeckNumber), props.Text{
					Top:   2,
					Size:  11,
					Align: align.Left,
				}),
			),
			col.New(3).Add(
				text.New("A2.Keys No.: "+strings.ToUpper(strconv.Itoa(insp.VehicleDets.NumberOfKeys)), props.Text{
					Top:   2,
					Size:  11,
					Align: align.Left,
				}),
			),
			col.New(4).Add(
				text.New("A3.Arrival Date: "+arrivalDate.Format(layout), props.Text{
					Top:   2,
					Size:  11,
					Align: align.Left,
				}),
			),
		),
		row.New(8).Add(
			col.New(6).Add(
				text.New("A4.Port of Discharge: DAR ES SALAAM PORT", props.Text{
					Top:   3,
					Size:  11,
					Align: align.Left,
				}),
			),
			col.New(3).Add(
				text.New("A5.Section:_______", props.Text{
					Top:   3,
					Size:  11,
					Align: align.Left,
				}),
			),
			col.New(4).Add(
				text.New("A6.Berth No: "+strings.ToUpper(manifest.BerthNo), props.Text{
					Top:   3,
					Size:  11,
					Align: align.Left,
				}),
			),
		),
	)

	// section B Ship's Particulars
	rows = append(rows,
		row.New(5),
		text.NewAutoRow("B. Ship's Particulars", props.Text{
			Size:  13,
			Style: fontstyle.Bold,
			Align: align.Left,
		}),

		row.New(7).Add(
			col.New(6).Add(
				text.New("B1.Ship's Name: "+strings.ToUpper(manifest.VesselName), props.Text{
					Top:   2,
					Size:  11,
					Align: align.Left,
				}),
			),
			col.New(3).Add(
				text.New("B2.Rot No:________", props.Text{
					Top:   2,
					Size:  11,
					Align: align.Left,
				}),
			),
			col.New(4).Add(
				text.New("B3.Voyage: "+strings.ToUpper(manifest.VoyageNo), props.Text{
					Top:   2,
					Size:  11,
					Align: align.Left,
				}),
			),
		),
	)

	// section C Vehicle Particulars

	rows = append(rows,
		row.New(5),
		text.NewAutoRow("C. Vehicle Particulars", props.Text{
			Size:  13,
			Style: fontstyle.Bold,
			Align: align.Left,
		}),
		row.New(2),
		row.New(10).Add(
			text.NewCol(4, "C1. Make: "+insp.VehicleDets.MakerName, props.Text{
				Top:   2.5,
				Left:  2,
				Size:  10,
				Align: align.Left,
			}).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full}),
			text.NewCol(4, "C4. Weight: "+strconv.FormatFloat(float64(insp.VehicleDets.Weight), 'f', 2, 64), props.Text{
				Top:   2.5,
				Left:  2,
				Size:  10,
				Align: align.Left,
			}).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full}),
			text.NewCol(4, "C7. B/Lading No:................", props.Text{
				Top:   2.5,
				Left:  2,
				Size:  10,
				Align: align.Left,
			}).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full}),
		),
		row.New(10).Add(
			text.NewCol(4, "C2. Type: "+insp.VehicleDets.BodyName, props.Text{
				Top:   2.5,
				Left:  2,
				Size:  10,
				Align: align.Left,
			}).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full}),
			text.NewCol(4, "C5. Volume: ", props.Text{
				Top:   2.5,
				Left:  2,
				Size:  10,
				Align: align.Left,
			}).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full}),
			text.NewCol(4, "C8. Destination: ........................", props.Text{
				Top:   2.5,
				Left:  2,
				Size:  10,
				Align: align.Left,
			}).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full}),
		),
		row.New(10).Add(
			text.NewCol(6, "C3. Chassis No: "+insp.VehicleDets.ChasisNumber, props.Text{
				Top:   2.5,
				Left:  2,
				Size:  10,
				Align: align.Left,
			}).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full}),
			text.NewCol(6, "C6. Port of of Loading: ............................", props.Text{
				Top:   2.5,
				Left:  2,
				Size:  10,
				Align: align.Left,
			}).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full}),
		),
	)

	rows = append(rows, row.New(5)) // spacer
	return rows
}

func buildMainfestSectionFault(manifest manifest.ManifestData) []core.Row {

	return []core.Row{
		row.New(5).Add(
			col.New(6).Add(
				text.New("Client Name: "+strings.ToUpper(manifest.ClientName), props.Text{
					Size:  13,
					Align: align.Left,
				}),
			),
			col.New(1),
			col.New(5).Add(
				text.New("Principal: MOL", props.Text{
					Size:  13,
					Align: align.Right,
				}),
			),
		),
		row.New(5),
	}
}

func getPageHeaderFault(manifest manifest.ManifestData) core.Row {
	imageBytes := specials.GetObjectBytes("https://openismila.s3.amazonaws.com/images/high/1744199842.webp")
	return row.New(20).Add(
		image.NewFromBytesCol(2, imageBytes, extension.Png, props.Rect{
			Center:             true,
			Percent:            100,
			JustReferenceWidth: true,
		}),
		col.New(3).Add(
			text.New("Walls", props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Left,
			}),
			text.New("International", props.Text{
				Top:   6,
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Left,
			}),
			text.New("Limited", props.Text{
				Top:   12,
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Left,
			}),
		),
		col.New(7).Add(
			text.New("Manifest No: "+strings.ToUpper(manifest.ManifestId[:8]), props.Text{
				Top:   2,
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Right,
				Color: getBlueColor(),
			}),
		),
	)
}

func getPageFooterFault() []core.Row {
	return []core.Row{
		row.New(5),
		row.New(5).Add(
			col.New(6).Add(
				text.New("WIL Surveyor: _____________________________________________", props.Text{
					// Top:   13,
					Style: fontstyle.BoldItalic,
					Size:  8,
					Align: align.Left,
				}),
			),
			col.New(6).Add(
				text.New("WIL Supervisor on behalf: _________________________________", props.Text{
					// Top:   13,
					Style: fontstyle.BoldItalic,
					Size:  8,
					Align: align.Right,
				}),
			),
		),
		row.New(8),
		row.New(5).Add(
			col.New(6).Add(
				text.New("Master/Ship's Officer: ______________________________________", props.Text{
					// Top:   13,
					Style: fontstyle.BoldItalic,
					Size:  8,
					Align: align.Left,
				}),
			),
			col.New(6).Add(
				text.New("Stamp and Signature: _____________________________________", props.Text{
					// Top:   13,
					Style: fontstyle.BoldItalic,
					Size:  8,
					Align: align.Right,
				}),
			),
		),
		row.New(9).Add(
			text.NewCol(12, "P.O. Box 62032. Email: info@walls.co.tz. Mobile: +255 786 833 930. Office: +255 763 175 801", props.Text{
				Size:  8,
				Align: align.Center,
				Top:   4,
			}),
		),
		row.New(5).Add(
			text.NewCol(12, "Palm Residence Building - Ocean Road, Third floor, Wing 3E, DSM Tanzania. Website www.walls.co.tz", props.Text{
				// Top:   8,
				Size:  8,
				Align: align.Center,
			}),
		),
	}
}
