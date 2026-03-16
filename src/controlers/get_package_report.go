package controlers

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/page"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/linestyle"
	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
	"github.com/shabs76/roro-local-server/specials"

	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/gin-gonic/gin"
)

func GeneratePackageReportPdf(c *gin.Context) {
	mId := c.Param("manifestId")
	// 1. Get manifest details
	st, manifests := manifestdataservices.SelectManifestInfo(" `manifest_id` = ? ", []any{mId})
	if st.State != constants.SuccessState {
		log.Printf("Error retrieving manifest information: %v", st.Data)
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

	// 2. Get inspected package details list
	st, packages := manifestdataservices.SelectPackageInspectionData(" manifest_packages.manifest_id = ?", []any{mId})
	if st.State != constants.SuccessState {
		log.Printf("Error fetching package inspection data: %s", st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": constants.ErrorState,
			"data":  "Failed to retrieve package inspection information",
		})
		return
	}

	// 3. Get uninspected packages details list
	st, uninspectedPackages := manifestdataservices.SelectPackageInfo(" manifest_id = ?  AND is_inspected != ? ", []any{mId, manifest.InspectionStatus.Yes})

	if st.State != constants.SuccessState {
		log.Printf("Error fetching uninspected packages: %s", st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": constants.ErrorState,
			"data":  "Failed to retrieve uninspected packages information",
		})
		return
	}

	if len(uninspectedPackages) == 0 && len(packages) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"state": constants.ErrorState,
			"data":  "No packages found",
		})
		return
	}

	// 4. Generate the PDF report
	m := getPackageReportMoroto(manifestDetails, packages, uninspectedPackages)

	document, err := m.Generate()
	if err != nil {
		log.Fatal(err.Error())
	}

	pdfBytes := document.GetBytes()

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=package_report.pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func getPackageReportMoroto(manifest manifest.ManifestData, inspectedPackages []manifest.PackageInspectionDetails, uninspectedPackages []manifest.PackageManifestInfo) core.Maroto {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	err := m.RegisterHeader(getPageHeaderPackage(manifest))
	if err != nil {
		log.Fatal(err.Error())
	}

	err = m.RegisterFooter(getPageFooterPackage()...)
	if err != nil {
		log.Fatal(err.Error())
	}

	m.AddRows(text.NewRow(16, "Package Discharge Survey Report (PDSR)", props.Text{
		Top:   5,
		Size:  16,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))
	m.AddRows(
		buildMainfestSectionPackage(manifest)...,
	)
	if len(inspectedPackages) > 0 {
		m.AddRows(
			buildInspectedPackageTable(inspectedPackages)...,
		)
	}

	// Insert a page break:
	newPage := page.New()

	if len(uninspectedPackages) > 0 {
		m.AddPages(newPage)
		m.AddRows(text.NewRow(16, "Uninspected Packages (Shortlandend Packages)", props.Text{
			Top:   5,
			Size:  16,
			Style: fontstyle.Bold,
			Align: align.Center,
		}))
		m.AddRows(
			buildMainfestSectionPackage(manifest)...,
		)
		m.AddRows(
			buildUninspectedPackageTable(uninspectedPackages)...,
		)
	}
	for _, pkg := range inspectedPackages {
		m.AddPages(newPage)
		m.AddRows(text.NewRow(16, "Package Inspection Details Report(PIDR)", props.Text{
			Top:   5,
			Size:  16,
			Style: fontstyle.Bold,
			Align: align.Center,
		}))
		m.AddRows(
			buildMainfestSectionPackage(manifest)...,
		)
		m.AddRows(buildEachPackageInspectionDetails(pkg)...)
	}

	return m
}

func buildEachPackageInspectionDetails(pkg manifest.PackageInspectionDetails) []core.Row {
	rows := []core.Row{}

	rows = append(rows, row.New(4))

	rows = append(rows, text.NewRow(7, "Package BOL No", props.Text{
		Size:  11.5,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))
	rows = append(rows, text.NewRow(7, pkg.BLNumber, props.Text{
		Size:  11.5,
		Style: fontstyle.Normal,
		Align: align.Left,
	}))
	rows = append(rows, text.NewRow(7, "Description", props.Text{
		Size:  11.5,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))
	rows = append(rows, text.NewRow(7, pkg.Description, props.Text{
		Size:  11.5,
		Style: fontstyle.Normal,
		Align: align.Left,
	}))
	rows = append(rows, text.NewRow(7, "Package Case", props.Text{
		Size:  11.5,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))
	rows = append(rows, text.NewRow(7, pkg.TypeName, props.Text{
		Size:  11.5,
		Style: fontstyle.Normal,
		Align: align.Left,
	}))

	rows = append(rows, row.New(4))

	imageBytes, err := specials.GetImageBytesForPdf(pkg.Picture)
	if err == nil {
		rows = append(rows, image.NewFromBytesRow(60, imageBytes, extension.Png, props.Rect{
			Center:             false,
			Percent:            100,
			JustReferenceWidth: true,
		}))
	} else {
		slog.Error(fmt.Sprintf("Error fetching/converting image for PDF: %s", err.Error()))
	}

	rows = append(rows, row.New(2))

	rows = append(rows, text.NewRow(7, "Inspection Status", props.Text{
		Size:  11,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))
	rows = append(rows, text.NewRow(7, pkg.InspectionStatus, props.Text{
		Size:  11,
		Style: fontstyle.Normal,
		Align: align.Left,
	}))
	rows = append(rows, text.NewRow(7, "Inspection Description", props.Text{
		Size:  11,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))
	rows = append(rows, text.NewRow(7, pkg.InspectionDescription, props.Text{
		Size:  11,
		Style: fontstyle.Normal,
		Align: align.Left,
	}))

	return rows
}

func buildUninspectedPackageTable(uninspectedPackages []manifest.PackageManifestInfo) []core.Row {
	rows := make([]core.Row, 0, len(uninspectedPackages)+2) // +1 for the header, +1 and +1 for the total row

	// body type table
	rows = append(rows, row.New(10).Add(
		text.NewCol(1, "SN", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.WhiteColor,
		}),
		text.NewCol(3, "Package BOL No", props.Text{
			Top:  2.5,
			Size: 10,

			Style: fontstyle.Bold,
			Align: align.Left,
			Right: 7,
			Color: &props.WhiteColor,
		}),
		text.NewCol(5, "Description", props.Text{
			Top:   2.5,
			Size:  10,
			Style: fontstyle.Bold,
			Align: align.Left,
			Right: 7,
			Color: &props.WhiteColor,
		}),
	).WithStyle(&props.Cell{BackgroundColor: getWallOrange()}))

	for i, pkg := range uninspectedPackages {
		if i%2 != 0 {
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1), props.Text{
					Top:   2.5,
					Left:  2,
					Size:  10,
					Align: align.Center,
					Color: &props.WhiteColor,
				}),
				text.NewCol(3, pkg.BLNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
					Color: &props.WhiteColor,
				}),
				text.NewCol(5, pkg.Description, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
					Color: &props.WhiteColor,
				}),
			).WithStyle(&props.Cell{BackgroundColor: getGrayColor()}))
		} else {
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1), props.Text{
					Top:   2.5,
					Left:  2,
					Size:  10,
					Align: align.Center,
				}),
				text.NewCol(3, pkg.BLNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
				}),
				text.NewCol(5, pkg.Description, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
				}),
			).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Bottom}))
		}
	}

	// total row
	rows = append(rows, row.New(10).Add(
		text.NewCol(5, "TOTAL", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &props.WhiteColor,
		}),
		text.NewCol(7, strconv.Itoa(len(uninspectedPackages)), props.Text{
			Top:   2.5,
			Size:  10,
			Style: fontstyle.Bold,
			Align: align.Right,
			Right: 7,
			Color: &props.WhiteColor,
		}),
	).WithStyle(&props.Cell{BackgroundColor: getWallOrange()}))

	return rows
}

func buildInspectedPackageTable(inspectedPackages []manifest.PackageInspectionDetails) []core.Row {
	rows := make([]core.Row, 0, len(inspectedPackages)+2) // +1 for the header, +1 and +1 for the total row

	// body type table
	rows = append(rows, row.New(10).Add(
		text.NewCol(1, "SN", props.Text{
			Top:   2.5,
			Size:  10,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &props.WhiteColor,
		}),
		text.NewCol(3, "BOL No", props.Text{
			Top:  2.5,
			Size: 10,

			Style: fontstyle.Bold,
			Align: align.Left,
			Right: 7,
			Color: &props.WhiteColor,
		}),
		text.NewCol(4, "Description", props.Text{
			Top:  2.5,
			Size: 10,

			Style: fontstyle.Bold,
			Align: align.Left,
			Right: 7,
			Color: &props.WhiteColor,
		}),
		text.NewCol(2, "Type", props.Text{
			Top:  2.5,
			Size: 10,

			Style: fontstyle.Bold,
			Align: align.Left,
			Right: 7,
			Color: &props.WhiteColor,
		}),
		text.NewCol(2, "Status", props.Text{
			Top:  2.5,
			Size: 10,

			Style: fontstyle.Bold,
			Align: align.Left,
			Right: 7,
			Color: &props.WhiteColor,
		}),
	).WithStyle(&props.Cell{BackgroundColor: getWallOrange()}))

	for i, pkg := range inspectedPackages {
		if i%2 != 0 {
			var dColor *props.Color = &props.WhiteColor
			if strings.ToLower(pkg.InspectionStatus) != "intact" {
				dColor = &props.RedColor
			}
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1), props.Text{
					Top:   2.5,
					Left:  2,
					Size:  10,
					Align: align.Center,
					Color: &props.WhiteColor,
				}),
				text.NewCol(3, pkg.BLNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
					Color: &props.WhiteColor,
				}),
				text.NewCol(4, pkg.Description, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
					Color: &props.WhiteColor,
				}),
				text.NewCol(2, pkg.TypeName, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
					Color: &props.WhiteColor,
				}),
				text.NewCol(2, pkg.InspectionStatus, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
					Color: dColor,
				}),
			).WithStyle(&props.Cell{BackgroundColor: getGrayColor()}))
		} else {
			var dColor *props.Color
			if strings.ToLower(pkg.InspectionStatus) != "intact" {
				dColor = getRedColor()
			}
			rows = append(rows, row.New(10).Add(
				text.NewCol(1, strconv.Itoa(i+1), props.Text{
					Top:   2.5,
					Left:  2,
					Size:  10,
					Align: align.Center,
				}),
				text.NewCol(3, pkg.BLNumber, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
				}),
				text.NewCol(4, pkg.Description, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
				}),
				text.NewCol(2, pkg.TypeName, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
				}),
				text.NewCol(2, pkg.InspectionStatus, props.Text{
					Top:   2.5,
					Size:  10,
					Align: align.Left,
					Color: dColor,
				}),
			).WithStyle(&props.Cell{BorderColor: getWallsGreen(), LineStyle: linestyle.Solid, BorderThickness: 0.1, BorderType: border.Bottom}))
		}
	}

	// total row
	rows = append(rows, row.New(10).Add(
		text.NewCol(6, "TOTAL", props.Text{
			Top:   2.5,
			Size:  12,
			Left:  3,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &props.WhiteColor,
		}),
		text.NewCol(6, strconv.Itoa(len(inspectedPackages)), props.Text{
			Top:  2.5,
			Size: 12,

			Style: fontstyle.Bold,
			Align: align.Right,
			Right: 7,
			Color: &props.WhiteColor,
		}),
	).WithStyle(&props.Cell{BackgroundColor: getWallOrange()}))

	return rows
}

func buildMainfestSectionPackage(manifest manifest.ManifestData) []core.Row {
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

func getPageHeaderPackage(manifest manifest.ManifestData) core.Row {
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

func getPageFooterPackage() []core.Row {
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
