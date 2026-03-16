package controlers

import (
	"log"
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
	"github.com/shabs76/roro-local-server/gendb"
	"github.com/shabs76/roro-local-server/specials"
)

func GenerateTallyReportPdf(c *gin.Context) {
	perPage := 250
	pg := c.Query("page")
	limit := c.DefaultQuery("limit", "all")
	ipg, erP := strconv.Atoi(pg)
	if erP != nil || ipg < 1 || limit == "all" {
		perPage = 10000
		ipg = 1
	}

	limitInt, erL := strconv.Atoi(limit)
	if erL == nil && limitInt > 0 && ipg >= 1 {
		perPage = limitInt
	}

	mId := c.Param("manifestId")

	// 1. Get manifest details
	st, manifests := manifestdataservices.SelectManifestInfo(" `manifest_id` = ? ", []any{mId})
	if st.State != constants.SuccessState {
		log.Printf("Error fetching manifest: %s", st.Data)
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

	manifestDetails := manifests[0]

	// 2. Get list of vehicle details
	subQr := " `manifest_id` = ? AND `inspection_status` = ? ORDER BY tallied_time ASC, inspection_time ASC "
	vals := []any{mId, "yes"}

	// pagination
	qrPage := "SELECT COUNT(vehicle_id) FROM `manifest_vehicles` WHERE " + subQr
	stx, rez := gendb.PagenationSelect(qrPage, ipg, perPage, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain pagination limit"})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": []any{}, "adv": 0, "per": 0})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(perPage)))

	st, vehicles := manifestdataservices.SelectVehicleInfo(subQr+" "+rez.Limit, vals)
	if st.State != constants.SuccessState {
		log.Println(st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "failed to obtain vehicles list"})
		return
	} else if len(vehicles) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"state": constants.ErrorState, "data": "No vehicles found"})
		return
	}

	// loop to check damage status on the vehicle
	overLanded := make([]manifest.VehiclesDetailsToShow, 0)
	withNoKeys := make([]manifest.VehiclesDetailsToShow, 0)
	vhShow := enrichVehicleDetails(vehicles)
	for i := range vhShow {
		if vhShow[i].IsDamaged == "yes" {
			overLanded = append(overLanded, vhShow[i])
		}
		if vhShow[i].IsOverLand == "yes" {
			overLanded = append(overLanded, vhShow[i])
		}
		if vhShow[i].NumberOfKeys == 0 {
			withNoKeys = append(withNoKeys, vhShow[i])
		}
	}

	// 3 Get Tally Summary
	st, tallySummary := manifestdataservices.GetCompleteSummary(mId)
	if st.State != constants.SuccessState {
		log.Printf("Error fetching tally summary: %s", st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": constants.ErrorState,
			"data":  "Failed to retrieve tally summary",
		})
		return
	}

	// 4 Get list of
	m := getTallyReportMoroto(vhShow, manifestDetails, tallySummary, overLanded, withNoKeys, perPage, ipg, totalPages)

	document, err := m.Generate()
	if err != nil {
		log.Fatal(err.Error())
	}

	pdfBytes := document.GetBytes()

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=tally_report.pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func getTallyReportMoroto(data []manifest.VehiclesDetailsToShow, manifest manifest.ManifestData, tallySummary manifest.VehicleSummary, overLand []manifest.VehiclesDetailsToShow, withNoKeys []manifest.VehiclesDetailsToShow, perPage, pg, totalPages int) core.Maroto {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	err := m.RegisterHeader(getPageHeaderTally(manifest))
	if err != nil {
		log.Fatal(err.Error())
	}

	err = m.RegisterFooter(getPageFooterTally()...)
	if err != nil {
		log.Fatal(err.Error())
	}

	if pg == 1 {
		m.AddRows(text.NewRow(16, "Vehicle Discharge Survey Report (VDSR)", props.Text{
			Top:   5,
			Size:  16,
			Style: fontstyle.Bold,
			Align: align.Center,
		}))

		m.AddRows(
			buildMainfestSectionTally(manifest)...,
		)
	}

	m.AddRows(
		buildTallyTable(data, pg, perPage)...,
	)

	// Insert a page break:
	newPage := page.New()
	if len(overLand) > 0 && pg == totalPages { // only add this section on the last page
		m.AddPages(newPage)

		m.AddRows(
			row.New(10),
			text.NewRow(14, "Overland Vehicle List", props.Text{
				Top:   2,
				Size:  15,
				Style: fontstyle.Bold,
				Align: align.Center,
			}))

		m.AddRows(
			buildMainfestSectionTally(manifest)...,
		)

		m.AddRows(
			buildOverLandedvehicle(overLand)...,
		)
	}

	// vehicles with no keys
	if len(withNoKeys) > 0 && pg == totalPages { // only add this section on the last page
		m.AddPages(newPage)
		m.AddRows(
			row.New(10),
			text.NewRow(14, "Vehicles With No Keys", props.Text{
				Top:   2,
				Size:  15,
				Style: fontstyle.Bold,
				Align: align.Center,
			}))

		m.AddRows(
			buildMainfestSectionTally(manifest)...,
		)

		m.AddRows(
			buildNoKeysTallyTable(withNoKeys)...,
		)
	}

	if pg == totalPages { // Only add these sections on the last page
		// Body types summary
		m.AddPages(newPage)
		m.AddRows(
			row.New(10),
			text.NewRow(14, "Vehicle Discharge Summary (Body Types)", props.Text{
				Top:   5,
				Size:  15,
				Style: fontstyle.Bold,
				Align: align.Center,
			}))
		m.AddRows(
			buildMainfestSectionTally(manifest)...,
		)
		m.AddRows(
			buildTallySummaryTableType(tallySummary)...,
		)

		// Makers summary
		m.AddPages(newPage)
		m.AddRows(
			row.New(10),

			text.NewRow(14, "Vehicle Discharge Summary (Makers)", props.Text{
				Top:   5,
				Size:  15,
				Style: fontstyle.Bold,
				Align: align.Center,
			}))
		m.AddRows(
			buildTallySummaryTableMaker(tallySummary)...,
		)

		// discharge quantity summary
		m.AddPages(newPage)
		m.AddRows(
			row.New(10),

			text.NewRow(14, "Discharge Quantity", props.Text{
				Top:   5,
				Size:  15,
				Style: fontstyle.Bold,
				Align: align.Center,
			}))
		m.AddRows(
			buildMainfestSectionTally(manifest)...,
		)
		m.AddRows(
			buildVehicleCountSummary(tallySummary)...,
		)
	}
	return m
}

func buildVehicleCountSummary(data manifest.VehicleSummary) []core.Row {
	rows := []core.Row{}

	// body type table
	rows = append(rows, row.New(10).Add(
		text.NewCol(9, "Name", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(3, "Quantity (Units)", props.Text{
			Top:   2.5,
			Size:  10,
			Style: fontstyle.Bold,
			Align: align.Right,
			Right: 7,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))
	rows = append(rows,
		row.New(9).Add(
			text.NewCol(9, "Given Chassis List", props.Text{
				Top:   2.5,
				Size:  10,
				Left:  3,
				Align: align.Left,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
			text.NewCol(3, strconv.Itoa(data.TotalCount), props.Text{
				Top:   2.5,
				Size:  10,
				Right: 7,
				Align: align.Right,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
		),
		row.New(10).Add(
			text.NewCol(9, "Vehicle Discharged", props.Text{
				Top:   2.5,
				Size:  10,
				Left:  3,
				Align: align.Left,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
			text.NewCol(3, strconv.Itoa(data.TotalCount), props.Text{
				Top:   2.5,
				Size:  10,
				Right: 7,
				Align: align.Right,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
		),
		row.New(10).Add(
			text.NewCol(9, "Vehicle Inspected", props.Text{
				Top:   2.5,
				Size:  10,
				Left:  3,
				Align: align.Left,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
			text.NewCol(3, strconv.Itoa(data.TotalCount), props.Text{
				Top:   2.5,
				Size:  10,
				Right: 7,
				Align: align.Right,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
		),
		row.New(10).Add(
			text.NewCol(9, "Over Landed", props.Text{
				Top:   2.5,
				Size:  10,
				Left:  3,
				Align: align.Left,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
			text.NewCol(3, strconv.Itoa(data.OverLand), props.Text{
				Top:   2.5,
				Size:  10,
				Right: 7,
				Align: align.Right,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
		),
		row.New(10).Add(
			text.NewCol(9, "Short Landed", props.Text{
				Top:   2.5,
				Size:  10,
				Left:  3,
				Align: align.Left,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
			text.NewCol(3, strconv.Itoa((data.TotalCount-data.Inspected)), props.Text{
				Top:   2.5,
				Size:  10,
				Right: 7,
				Align: align.Right,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full, BackgroundColor: &props.WhiteColor,
			}),
		),
	)
	return rows
}

func buildTallySummaryTableType(data manifest.VehicleSummary) []core.Row {
	rows := make([]core.Row, 0, len(data.BodyTypes)+2) // +1 for the header, +1 and +1 for the total row

	// body type table
	rows = append(rows, row.New(10).Add(
		text.NewCol(9, "Body Type", props.Text{
			Top:   2.5,
			Size:  12,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(3, "Quantity (Units)", props.Text{
			Top:   2.5,
			Size:  12,
			Style: fontstyle.Bold,
			Align: align.Right,
			Right: 7,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))
	for _, body := range data.BodyTypes {
		rows = append(rows, row.New(10).Add(
			text.NewCol(9, body.BodyName, props.Text{
				Top:   2.5,
				Size:  10,
				Left:  3,
				Align: align.Left,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
			}),
			text.NewCol(3, strconv.Itoa(body.Count), props.Text{
				Top:   2.5,
				Size:  10,
				Right: 7,
				Align: align.Right,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
			}),
		))
	}
	// add total row
	rows = append(rows, row.New(10).Add(
		text.NewCol(9, "Total", props.Text{
			Top:   2.5,
			Size:  12,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(3, strconv.Itoa(data.TotalCount), props.Text{
			Top:   2.5,
			Size:  12,
			Style: fontstyle.Bold,
			Align: align.Right,
			Right: 7,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))
	return rows
}

func buildTallySummaryTableMaker(data manifest.VehicleSummary) []core.Row {
	rows := make([]core.Row, 0, len(data.Makers)+2) // +1 for the header, +1 and +1 for the total row
	// maker table
	rows = append(rows, row.New(10).Add(
		text.NewCol(9, "Maker", props.Text{
			Top:   2.5,
			Size:  12,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(3, "Quantity (Units)", props.Text{
			Top:   2.5,
			Size:  12,
			Style: fontstyle.Bold,
			Align: align.Right,
			Right: 7,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))
	for _, maker := range data.Makers {
		rows = append(rows, row.New(10).Add(
			text.NewCol(9, maker.MakerName, props.Text{
				Top:   2.5,
				Size:  10,
				Left:  3,
				Align: align.Left,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
			}),
			text.NewCol(3, strconv.Itoa(maker.Count), props.Text{
				Top:   2.5,
				Size:  10,
				Right: 7,
				Align: align.Right,
			}).WithStyle(&props.Cell{
				BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
			}),
		))
	}
	// add total row
	rows = append(rows, row.New(10).Add(
		text.NewCol(9, "Total", props.Text{
			Top:   2.5,
			Size:  12,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(3, strconv.Itoa(data.TotalCount), props.Text{
			Top:   2.5,
			Size:  12,
			Right: 7,
			Style: fontstyle.Bold,
			Align: align.Right,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))

	return rows
}

func buildOverLandedvehicle(data []manifest.VehiclesDetailsToShow) []core.Row {
	rows := []core.Row{} // +1 for the header row
	// create table header
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
		text.NewCol(4, "Chassis No", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(4, "Maker & Model", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(2, "Body Type", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),

		text.NewCol(1, "Deck", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))

	for i, vehicle := range data {
		model := ""
		if vehicle.VehicleModel != "" && strings.ToLower(strings.TrimSpace(vehicle.VehicleModel)) != "notset" {
			model = vehicle.VehicleModel
		}
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
				text.NewCol(4, vehicle.ChasisNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(4, vehicle.Maker+", "+model, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(2, vehicle.BodyType, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),

				text.NewCol(1, vehicle.DeckNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
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
				text.NewCol(4, vehicle.ChasisNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}),
				text.NewCol(4, vehicle.Maker+", "+model, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(2, vehicle.BodyType, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),

				text.NewCol(1, vehicle.DeckNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
			))
		}
	}

	return rows
}

func buildNoKeysTallyTable(data []manifest.VehiclesDetailsToShow) []core.Row {
	rows := []core.Row{} // +1 for the header row
	// create table header
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
		text.NewCol(4, "Chassis No", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(4, "Maker & Model", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(2, "Body Type", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),

		text.NewCol(1, "Deck", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))

	for i, vehicle := range data {
		model := ""
		if vehicle.VehicleModel != "" && strings.ToLower(strings.TrimSpace(vehicle.VehicleModel)) != "notset" {
			model = vehicle.VehicleModel
		}
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
				text.NewCol(4, vehicle.ChasisNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(4, vehicle.Maker+", "+model, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(2, vehicle.BodyType, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),

				text.NewCol(1, vehicle.DeckNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
			))
		} else {
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1), props.Text{
					Top:  2.5,
					Left: 2,
					Size: 10,

					Align: align.Center,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(4, vehicle.ChasisNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(4, vehicle.Maker+", "+model, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(2, vehicle.BodyType, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),

				text.NewCol(1, vehicle.DeckNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
			))
		}
	}

	return rows
}

func buildTallyTable(data []manifest.VehiclesDetailsToShow, pg, perPage int) []core.Row {
	rows := []core.Row{} // +1 for the header row
	// create table header
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
		text.NewCol(3, "Chassis No", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(4, "Maker & Model", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(2, "Body Type", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),

		text.NewCol(1, "Deck", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
		text.NewCol(1, "Damage", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  1,
			Style: fontstyle.Bold,
			Align: align.Left,
		}).WithStyle(&props.Cell{
			BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.3, BorderType: border.Full,
		}),
	))

	// get what to add to sn to get the real number
	startIndex := (pg - 1) * perPage

	for i, vehicle := range data {
		model := ""
		if vehicle.VehicleModel != "" && strings.ToLower(strings.TrimSpace(vehicle.VehicleModel)) != "notset" {
			model = vehicle.VehicleModel
		}
		if i%2 != 0 {
			var dColor *props.Color
			if vehicle.IsDamaged == "yes" {
				dColor = &props.RedColor
			}
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1+startIndex), props.Text{
					Top:   2.5,
					Left:  2,
					Size:  10,
					Align: align.Center,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(3, vehicle.ChasisNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(4, vehicle.Maker+", "+model, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(2, vehicle.BodyType, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),

				text.NewCol(1, vehicle.DeckNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(1, vehicle.IsDamaged, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  1,
					Align: align.Left,
					Color: dColor,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
			))
		} else {
			var dColor *props.Color
			if vehicle.IsDamaged == "yes" {
				dColor = &props.RedColor
			}
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1+startIndex), props.Text{
					Top:   2.5,
					Left:  2,
					Size:  10,
					Align: align.Center,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(3, vehicle.ChasisNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(4, vehicle.Maker+", "+model, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(2, vehicle.BodyType, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),

				text.NewCol(1, vehicle.DeckNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  3,
					Align: align.Left,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
				text.NewCol(1, vehicle.IsDamaged, props.Text{
					Top:   2.5,
					Size:  10,
					Left:  1,
					Align: align.Left,
					Color: dColor,
				}).WithStyle(&props.Cell{
					BorderColor: &props.BlackColor, LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Full,
				}),
			))
		}
	}

	return rows
}

func buildMainfestSectionTally(manifest manifest.ManifestData) []core.Row {
	now := time.Now()
	layout := "Mon, 2 Jan 2006"

	return []core.Row{
		row.New(5).Add(
			col.New(6).Add(
				text.New("Client Name: "+strings.ToUpper(manifest.ClientName), props.Text{
					Size:  11,
					Align: align.Left,
				}),
			),
			col.New(1),
			col.New(5).Add(
				text.New("Principal: MOL", props.Text{
					Size:  11,
					Align: align.Right,
				}),
			),
		),
		row.New(7).Add(
			col.New(4).Add(
				text.New("Vessel: "+strings.ToUpper(manifest.VesselName), props.Text{
					Size:  11,
					Top:   2,
					Align: align.Left,
				}),
			),
			col.New(1),
			col.New(4).Add(
				text.New("Voyage: "+strings.ToUpper(manifest.VoyageNo), props.Text{
					Top:   2,
					Size:  11,
					Align: align.Left,
				}),
			),
			col.New(3).Add(
				text.New("Date: "+now.Format(layout), props.Text{
					Top:   2,
					Size:  11,
					Align: align.Right,
				}),
			),
		),
		row.New(3),
	}
}

func getPageHeaderTally(manifest manifest.ManifestData) core.Row {
	imageBytes := specials.GetObjectBytes("https://openismila.s3.amazonaws.com/images/high/1744199842.webp")
	return row.New(20).Add(
		image.NewFromBytesCol(2, imageBytes, extension.Png, props.Rect{
			Center:  true,
			Percent: 100,
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

func getPageFooterTally() []core.Row {
	return []core.Row{
		row.New(5),
		row.New(5).Add(
			col.New(5).Add(
				text.New("WIL Surveyor: ____________________________________", props.Text{
					// Top:   13,
					Style: fontstyle.BoldItalic,
					Size:  8,
					Align: align.Left,
				}),
			),
			col.New(1),
			col.New(6).Add(
				text.New("Master/Ship's Officer: ____________________________________", props.Text{
					// Top:   13,
					Style: fontstyle.BoldItalic,
					Size:  8,
					Align: align.Left,
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
