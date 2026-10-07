package controlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	apiservices "github.com/shabs76/roro-local-server/api_services"
	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
	"github.com/shabs76/roro-local-server/pkg/middleware"
	"github.com/shabs76/roro-local-server/specials"
)

// A publish run uploads inspected vehicles and packages of one manifest to the remote
// server. It runs in the background, independent of the SSE request that started it:
// a tablet that sleeps or loses Wi-Fi no longer stops the run, and a second trigger
// for the same manifest attaches to the running job instead of uploading the same
// items twice.

type publishEvent struct {
	Event string
	Data  any
}

type publishFailure struct {
	Kind   string `json:"kind"` // vehicle or package
	Id     string `json:"id"`
	Label  string `json:"label"` // chassis number or package number
	Reason string `json:"reason"`
}

type publishJob struct {
	manifestId string

	mu          sync.Mutex
	subs        map[chan publishEvent]struct{}
	total       int
	completed   int
	published   int
	failedItems []publishFailure
	warnings    []string
	fatal       string // set when the run could not start

	done chan struct{}
}

var (
	publishJobsMu sync.Mutex
	publishJobs   = map[string]*publishJob{}
)

var errMediaFileMissing = errors.New("file missing on server")

// runPublishSSE starts (or attaches to) the publish job of a manifest and streams its
// progress. addedLaterOnly limits the run to vehicles and packages added on site.
func runPublishSSE(c *gin.Context, addedLaterOnly bool) {
	val, ok := c.Get("authSession")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}
	authSession, ok := val.(middleware.AuthSession)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	manifestId := c.Param("manifestId")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": "error", "data": "Missing manifest ID"})
		return
	}
	// "continue" (default) publishes only items that are not published yet; any other
	// value publishes every inspected item again.
	isContinue := c.DefaultQuery("status", "continue") == "continue"

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	publishJobsMu.Lock()
	job, running := publishJobs[manifestId]
	if !running {
		job = &publishJob{manifestId: manifestId, subs: map[chan publishEvent]struct{}{}, done: make(chan struct{})}
		publishJobs[manifestId] = job
	}
	ch := job.subscribe()
	publishJobsMu.Unlock()
	defer job.unsubscribe(ch)

	if running {
		c.SSEvent("progress", progressData("All", "A publish for this manifest is already running. Showing its progress.", job.progressPercent()))
	} else {
		go job.run(authSession, isContinue, addedLaterOnly)
		c.SSEvent("progress", progressData("All", "Publish started", 0))
	}
	c.Writer.Flush()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-ch:
			c.SSEvent(msg.Event, msg.Data)
			c.Writer.Flush()
		case <-job.done:
			// Deliver what is still queued, then the summary. The summary is built here
			// from the job state, so it can never be dropped.
			for {
				select {
				case msg := <-ch:
					c.SSEvent(msg.Event, msg.Data)
					continue
				default:
				}
				break
			}
			c.SSEvent("complete", job.completeData())
			c.Writer.Flush()
			return
		case <-ticker.C:
			c.Writer.WriteString(": keep-alive\n\n")
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			// The tablet went away; the job keeps running.
			return
		}
	}
}

func progressData(task, message string, progress int) gin.H {
	return gin.H{"state": "success", "data": gin.H{"task": task, "message": message, "status": "running", "progress": progress}}
}

func (j *publishJob) subscribe() chan publishEvent {
	ch := make(chan publishEvent, 256)
	j.mu.Lock()
	j.subs[ch] = struct{}{}
	j.mu.Unlock()
	return ch
}

func (j *publishJob) unsubscribe(ch chan publishEvent) {
	j.mu.Lock()
	delete(j.subs, ch)
	j.mu.Unlock()
}

// broadcast never blocks: a slow or gone listener misses progress lines but still
// gets the final summary.
func (j *publishJob) broadcast(ev publishEvent) {
	j.mu.Lock()
	defer j.mu.Unlock()
	for ch := range j.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

func (j *publishJob) progressPercent() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.total == 0 {
		return 0
	}
	p := j.completed * 100 / j.total
	if p >= 100 {
		p = 99
	}
	return p
}

func (j *publishJob) progress(task, message string) {
	j.broadcast(publishEvent{Event: "progress", Data: progressData(task, message, j.progressPercent())})
}

func (j *publishJob) fail(kind, id, label, reason string) {
	slog.Error("Publish item failed", "manifest", j.manifestId, "kind", kind, "id", id, "label", label, "reason", reason)
	j.mu.Lock()
	j.failedItems = append(j.failedItems, publishFailure{Kind: kind, Id: id, Label: label, Reason: reason})
	j.mu.Unlock()
	j.broadcast(publishEvent{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to publish %s %s: %s", kind, label, reason)}})
}

func (j *publishJob) warn(message string) {
	slog.Warn("Publish warning", "manifest", j.manifestId, "message", message)
	j.mu.Lock()
	j.warnings = append(j.warnings, message)
	j.mu.Unlock()
	j.broadcast(publishEvent{Event: "warning", Data: gin.H{"state": "warning", "message": message}})
}

func (j *publishJob) itemDone(published bool) {
	j.mu.Lock()
	j.completed++
	if published {
		j.published++
	}
	j.mu.Unlock()
}

func (j *publishJob) completeData() gin.H {
	j.mu.Lock()
	defer j.mu.Unlock()

	state := "success"
	message := fmt.Sprintf("Upload completed: %d of %d published", j.published, j.total)
	switch {
	case j.fatal != "":
		state = "error"
		message = j.fatal
	case j.total == 0:
		message = "Nothing to publish"
	case len(j.failedItems) > 0:
		state = "partial"
		message = fmt.Sprintf("Upload finished with errors: %d of %d published, %d failed", j.published, j.total, len(j.failedItems))
	}

	failed := j.failedItems
	if failed == nil {
		failed = []publishFailure{}
	}
	warnings := j.warnings
	if warnings == nil {
		warnings = []string{}
	}

	return gin.H{
		"state": state,
		"data": gin.H{
			"task":        "All",
			"message":     message,
			"status":      "finished",
			"progress":    100,
			"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
			"total":       j.total,
			"published":   j.published,
			"failed":      len(j.failedItems),
			"failedItems": failed,
			"warnings":    warnings,
		},
	}
}

func envInt(name string, def int) int {
	if v, err := strconv.Atoi(specials.GetEnvVariable(name, "")); err == nil && v > 0 {
		return v
	}
	return def
}

func (j *publishJob) run(auth middleware.AuthSession, isContinue, addedLaterOnly bool) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Publish job panicked", "manifest", j.manifestId, "panic", r)
			j.mu.Lock()
			j.fatal = fmt.Sprintf("Internal error: %v", r)
			j.mu.Unlock()
		}
		publishJobsMu.Lock()
		delete(publishJobs, j.manifestId)
		publishJobsMu.Unlock()
		close(j.done)
	}()

	started := time.Now()

	pkgSubQuery := " manifest_packages.manifest_id = ? "
	vehSubQuery := " manifest_vehicles.manifest_id = ? AND manifest_vehicles.inspection_status = 'yes' "
	if addedLaterOnly {
		pkgSubQuery += " AND manifest_packages.is_added_later = 'yes' "
		vehSubQuery += " AND manifest_vehicles.is_added_later = 'yes' "
	}
	if isContinue {
		pkgSubQuery += " AND manifest_packages.is_published != 'yes' "
		vehSubQuery += " AND manifest_vehicles.is_published != 'yes' "
	}

	stP, packages := manifestdataservices.SelectPackageInspectionData(pkgSubQuery, []any{j.manifestId})
	if stP.State != constants.SuccessState {
		j.mu.Lock()
		j.fatal = "Failed to fetch packages"
		j.mu.Unlock()
		return
	}
	stV, vehicles := manifestdataservices.SelectVehicleAndTallyDetails(vehSubQuery, []any{j.manifestId})
	if stV.State != constants.SuccessState {
		j.mu.Lock()
		j.fatal = "Failed to fetch vehicles"
		j.mu.Unlock()
		return
	}

	j.mu.Lock()
	j.total = len(packages) + len(vehicles)
	j.mu.Unlock()

	uploader := &mediaUploader{
		auth:  auth,
		sem:   make(chan struct{}, envInt("PUBLISH_MEDIA_WORKERS", 6)),
		files: map[string]*mediaUpload{},
	}

	// A small worker pool publishes several items at once. Media uploads of all items
	// share the uploader's limit, so the remote server sees bounded parallel load.
	tasks := make(chan func())
	var wg sync.WaitGroup
	for i := 0; i < envInt("PUBLISH_ITEM_WORKERS", 3); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				task()
			}
		}()
	}
	for _, pkg := range packages {
		pkg := pkg
		tasks <- func() { j.itemDone(j.publishPackage(uploader, auth, pkg)) }
	}
	for _, veh := range vehicles {
		veh := veh
		tasks <- func() { j.itemDone(j.publishVehicle(uploader, auth, veh)) }
	}
	close(tasks)
	wg.Wait()

	slog.Info("Publish finished", "manifest", j.manifestId, "total", j.total, "published", j.published,
		"failed", len(j.failedItems), "duration", time.Since(started).String())
}

func (j *publishJob) publishPackage(up *mediaUploader, auth middleware.AuthSession, pkg manifest.PackageInspectionDetails) bool {
	label := pkg.PackageNumber
	task := "Package Upload"
	msg := fmt.Sprintf("Processing Package %s [%s]", pkg.PackageNumber, pkg.Description)

	if pkg.IsAddedLater == "yes" {
		j.progress(task, msg+": Registering New Package")
		newPkgId, err := apiservices.UploadSinglePackageToRemote(auth.LogID, auth.LogKey, manifest.AddPackageRequest{
			ManifestId:    j.manifestId,
			BLNumber:      pkg.BLNumber,
			PackageNumber: pkg.PackageNumber,
			Description:   pkg.Description,
		})
		if err != nil {
			j.fail("package", pkg.PackageId, label, "register new package: "+err.Error())
			return false
		}
		if st := manifestdataservices.UpdatePackageIdToRemote(pkg.PackageId, newPkgId); st.State != constants.SuccessState {
			j.fail("package", pkg.PackageId, label, st.Data)
			return false
		}
		pkg.PackageId = newPkgId
	}

	j.progress(task, msg+": Uploading media")
	paths := []string{pkg.Picture}
	for _, m := range pkg.Media {
		paths = append(paths, m.MediaLink)
	}
	urls, errs := up.uploadAll(paths)

	if err := errs[pkg.Picture]; err != nil {
		j.fail("package", pkg.PackageId, label, "package image: "+err.Error())
		return false
	}
	medias := []manifest.PackageExtraMediaRequest{}
	for _, m := range pkg.Media {
		if m.MediaLink == "" {
			continue
		}
		if err := errs[m.MediaLink]; err != nil {
			if errors.Is(err, errMediaFileMissing) {
				j.warn(fmt.Sprintf("Package %s: extra media skipped, %v", label, err))
				continue
			}
			j.fail("package", pkg.PackageId, label, "extra media: "+err.Error())
			return false
		}
		medias = append(medias, manifest.PackageExtraMediaRequest{MediaLink: urls[m.MediaLink], MediaType: m.MediaType, Remark: m.Remark})
	}

	j.progress(task, msg+": Saving Data")
	_, err := apiservices.UploadPackageInspectionData(auth.LogID, auth.LogKey, manifest.PackageInspectionSaveRequest{
		PackageId:          pkg.PackageId,
		PackageImage:       urls[pkg.Picture],
		TypeId:             pkg.TypeId,
		InspectionTime:     pkg.InspectionTime,
		InspectionStatusId: pkg.InspectionStatusId,
		Media:              medias,
	})
	if err != nil {
		j.fail("package", pkg.PackageId, label, err.Error())
		return false
	}

	if st := manifestdataservices.UpdatePackagePublishedStatus(pkg.PackageId, true); st.State != constants.SuccessState {
		j.fail("package", pkg.PackageId, label, "published on remote, but the local published flag was not saved: "+st.Data)
		return false
	}
	manifestdataservices.DeleteMediaRemoteUrls(paths)
	return true
}

func (j *publishJob) publishVehicle(up *mediaUploader, auth middleware.AuthSession, veh manifest.VehiclesDetailsAndTally) bool {
	label := veh.ChasisNumber
	task := "Vehicle Upload"
	msg := fmt.Sprintf("Processing Vehicle %s [%s]", veh.ChasisNumber, veh.Description)

	if veh.IsAddedLater == "yes" {
		j.progress(task, msg+": Registering New Vehicle")
		isOverLand := "no"
		if veh.OverLandStatus == "yes" {
			isOverLand = "yes"
		}
		newVehId, err := apiservices.UploadSingleVehicleToRemote(auth.LogID, auth.LogKey, manifest.AddVehicleRequest{
			ManifestId:   j.manifestId,
			ChasisNumber: veh.ChasisNumber,
			VehicleModel: veh.VehicleModel,
			Description:  veh.Description,
			Weight:       float64(veh.Weight),
			IsOverLand:   isOverLand,
			BLNumber:     veh.BLNumber,
		})
		if err != nil {
			j.fail("vehicle", veh.VehicleId, label, "register new vehicle: "+err.Error())
			return false
		}
		if st := manifestdataservices.UpdateVehicleIdToRemote(veh.VehicleId, newVehId); st.State != constants.SuccessState {
			j.fail("vehicle", veh.VehicleId, label, st.Data)
			return false
		}
		veh.VehicleId = newVehId
	}

	j.progress(task, msg+": Reading inspection")
	stc, checks := manifestdataservices.SelectInspectionChecksForPublish(veh.VehicleId)
	str, remarks := manifestdataservices.SelectInspectionRemarks(" vehicle_id = ? ", []any{veh.VehicleId})
	stm, gallery := manifestdataservices.SelectVehicleMedia(" vehicle_id = ? ", []any{veh.VehicleId})
	sto, onBoard := manifestdataservices.SelectOnBoardPackage(" vehicle_id = ? ", []any{veh.VehicleId})
	for _, st := range []*constants.AnswerState{stc, str, stm, sto} {
		if st.State != constants.SuccessState {
			j.fail("vehicle", veh.VehicleId, label, "read local inspection data: "+st.Data)
			return false
		}
	}

	// Upload every media file of the vehicle in parallel.
	paths := []string{veh.VehicleImage}
	for _, c := range checks {
		paths = append(paths, c.FaultImage)
	}
	for _, r := range remarks {
		paths = append(paths, r.ImageLink)
	}
	for _, m := range gallery {
		paths = append(paths, m.MediaLink)
	}
	for _, o := range onBoard {
		for _, om := range o.Media {
			paths = append(paths, om.MediaLink)
		}
	}
	j.progress(task, msg+": Uploading media")
	urls, errs := up.uploadAll(paths)

	// The main image is required. Other media whose file is missing on this server
	// cannot be fixed by a retry, so they are skipped with a warning; any other upload
	// error fails the vehicle so the next publish retries it.
	if err := errs[veh.VehicleImage]; err != nil {
		reason := "vehicle image: " + err.Error()
		if errors.Is(err, errMediaFileMissing) {
			reason = "image file missing on server (" + veh.VehicleImage + ")"
		}
		j.fail("vehicle", veh.VehicleId, label, reason)
		return false
	}
	media := func(path, what string) (string, bool) {
		if path == "" {
			return "", true
		}
		if err := errs[path]; err != nil {
			if errors.Is(err, errMediaFileMissing) {
				j.warn(fmt.Sprintf("Vehicle %s: %s skipped, %v", label, what, err))
				return "", true
			}
			j.fail("vehicle", veh.VehicleId, label, what+": "+err.Error())
			return "", false
		}
		return urls[path], true
	}

	reqChecks := []manifest.InspectionCheck{}
	for _, c := range checks {
		url, ok := media(c.FaultImage, "fault image")
		if !ok {
			return false
		}
		reqChecks = append(reqChecks, manifest.InspectionCheck{CheckId: c.CheckId, CheckStatus: c.Status, FaultImage: url})
	}
	reqRemarks := []manifest.RemarkSaveRequest{}
	for _, r := range remarks {
		url, ok := media(r.ImageLink, "remark image")
		if !ok {
			return false
		}
		reqRemarks = append(reqRemarks, manifest.RemarkSaveRequest{Remark: r.Remark, RemarkType: r.RemarkType, RemarkImage: url})
	}
	reqMedia := []manifest.VehicleExtraMediaRequest{}
	for _, m := range gallery {
		if m.MediaLink == "" {
			continue
		}
		url, ok := media(m.MediaLink, "extra media")
		if !ok {
			return false
		}
		if url == "" {
			continue
		}
		reqMedia = append(reqMedia, manifest.VehicleExtraMediaRequest{MediaLink: url, MediaType: m.MediaType, Remark: m.Remark})
	}
	reqPackages := []manifest.OnBoardPackageRequest{}
	for _, o := range onBoard {
		obMedia := []manifest.OnBoardPackageMediaRequest{}
		for _, om := range o.Media {
			if om.MediaLink == "" {
				continue
			}
			url, ok := media(om.MediaLink, "on-board package media")
			if !ok {
				return false
			}
			if url == "" {
				continue
			}
			obMedia = append(obMedia, manifest.OnBoardPackageMediaRequest{MediaLink: url, MediaType: om.MediaType})
		}
		reqPackages = append(reqPackages, manifest.OnBoardPackageRequest{Title: o.Title, Remark: o.Remark, Media: obMedia})
	}

	j.progress(task, msg+": Saving Data")
	_, err := apiservices.UploadVehicleInspectionData(auth.LogID, auth.LogKey, manifest.InspectionChecksRequest{
		VehicleId:      veh.VehicleId,
		VehicleImage:   urls[veh.VehicleImage],
		ManifestId:     j.manifestId,
		MakerId:        veh.MakerId,
		BodyId:         veh.BodyId,
		ModelName:      veh.VehicleModel,
		InspectionTime: veh.InspectionTime,
		// The tablet's id of this inspection: a later publish of the same inspection is
		// then a resend for the remote server, not a re-inspection.
		SubmissionId: manifestdataservices.SelectTallySubmissionId(veh.VehicleId),
		DeckNumber:   veh.DeckNumber,
		NumberOfKeys: veh.NumberOfKeys,
		KeyType:      veh.KeyType,
		Checks:       reqChecks,
		Remarks:      reqRemarks,
		Media:        reqMedia,
		Packages:     reqPackages,
	})
	if err != nil {
		j.fail("vehicle", veh.VehicleId, label, err.Error())
		return false
	}

	if st := manifestdataservices.UpdateVehiclePublishedStatus(veh.VehicleId, true); st.State != constants.SuccessState {
		j.fail("vehicle", veh.VehicleId, label, "published on remote, but the local published flag was not saved: "+st.Data)
		return false
	}
	manifestdataservices.DeleteMediaRemoteUrls(paths)
	vehicleEvents.notify(j.manifestId, veh.VehicleId, "published")
	return true
}

// mediaUploader uploads local media files to the remote server with bounded
// parallelism. Each file is uploaded at most once per run, even when several
// vehicles use it at the same time.
type mediaUploader struct {
	auth middleware.AuthSession
	sem  chan struct{}

	mu    sync.Mutex
	files map[string]*mediaUpload
}

type mediaUpload struct {
	once sync.Once
	url  string
	err  error
}

// uploadAll uploads the given files in parallel. It returns the remote URL and the
// error of each non-empty path.
func (u *mediaUploader) uploadAll(paths []string) (map[string]string, map[string]error) {
	urls := map[string]string{}
	errs := map[string]error{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	seen := map[string]bool{}
	for _, p := range paths {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			url, err := u.upload(p)
			mu.Lock()
			if err != nil {
				errs[p] = err
			} else {
				urls[p] = url
			}
			mu.Unlock()
		}(p)
	}
	wg.Wait()
	return urls, errs
}

func (u *mediaUploader) upload(localPath string) (string, error) {
	u.mu.Lock()
	f, ok := u.files[localPath]
	if !ok {
		f = &mediaUpload{}
		u.files[localPath] = f
	}
	u.mu.Unlock()

	f.once.Do(func() { f.url, f.err = u.uploadFile(localPath) })
	return f.url, f.err
}

func (u *mediaUploader) uploadFile(localPath string) (string, error) {
	// Uploaded by an earlier run that did not finish: reuse the remote copy.
	if url, ok := manifestdataservices.SelectMediaRemoteUrl(localPath); ok {
		return url, nil
	}

	finalPath, ok := resolveMediaPath(localPath)
	if !ok {
		return "", fmt.Errorf("%w: %s", errMediaFileMissing, localPath)
	}

	u.sem <- struct{}{}
	defer func() { <-u.sem }()

	var url string
	var err error
	switch strings.ToLower(filepath.Ext(finalPath)) {
	case ".mp4", ".avi", ".mov", ".mkv":
		url, err = apiservices.UploadVideo(u.auth.LogID, u.auth.LogKey, finalPath)
	case ".pdf", ".doc", ".docx":
		url, err = apiservices.UploadPdf(u.auth.LogID, u.auth.LogKey, finalPath)
	default:
		// Images, and anything unknown, go to the image endpoint as before.
		url, err = apiservices.UploadImage(u.auth.LogID, u.auth.LogKey, finalPath)
	}
	if err != nil {
		return "", err
	}

	manifestdataservices.SaveMediaRemoteUrl(localPath, url)
	return url, nil
}

// resolveMediaPath finds a stored media file. Paths come from the tablets as
// "images/x.jpg" or "media_data/images/x.jpg", relative to the server or to src/.
func resolveMediaPath(localPath string) (string, bool) {
	cleanPath := strings.TrimPrefix(localPath, "/")
	candidates := []string{cleanPath, filepath.Join(constants.MediaBaseDir, cleanPath)}
	for _, p := range candidates[:2] {
		candidates = append(candidates, filepath.Join("..", p))
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, true
		}
	}
	return "", false
}
