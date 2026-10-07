# Tablet app: saving inspections and keeping vehicle status in sync

This page explains how the Flutter tablet app should use the local server so that:

- a retried save never shows up as a second inspection,
- every tablet sees, within seconds, the vehicles that other tablets inspected,
- publish results are shown correctly.

All endpoints need the usual `Log-Id` and `Log-Key` headers.

## 1. Saving an inspection

`POST /manifest/inspection/save/vehicle/inspection`

The body is the same as before, plus two optional fields:

| Field | Type | Meaning |
|---|---|---|
| `submissionId` | string, up to 64 chars | A UUID v4. Create it **once**, when the inspection form opens. Store it with the queued payload and send the **same** value on every retry. |
| `reinspect` | bool | Send `false` for a normal inspection. Send `true` only after the user has confirmed that they want to replace an existing inspection. If you leave it out, the server keeps the old behaviour (it accepts the re-inspection). |

The server matches retries in this order:

1. If `submissionId` matches the stored one, the request is a resend.
2. Older builds send no `submissionId`. For them, the same `inspectionTime` from the same user counts as a resend.

A resend is applied in place, so it never creates a history entry.

### Responses

| HTTP | Body | What the app does |
|---|---|---|
| 200 | `{"state":"success","adv":"new"}` | Saved. Remove it from the queue. |
| 200 | `{"state":"success","adv":"resend"}` | Already saved by an earlier try. Remove it from the queue. |
| 200 | `{"state":"success","adv":"reinspection"}` | The previous inspection was archived and replaced. Remove it from the queue. |
| 409 | `{"state":"conflict","data":{"inspectedBy":"…","inspectionTime":"…","message":"…"}}` | Another inspection exists, and `reinspect` was `false`. **Stop retrying.** Show "Inspected by X at T. Re-inspect?". If the user confirms, send again with `reinspect: true` and the same `submissionId`. |
| 400 | `{"state":"error",…}` | The payload is invalid. **Stop retrying.** Show the error. |
| 404 | `{"state":"error","adv":"not_found",…}` | The server does not know this vehicle (wrong manifest, or an added-later vehicle not registered). **Stop retrying.** |
| 422 | `{"state":"error","adv":"invalid_reference",…}` | The inspection names a check, maker or body type the server does not have. **Stop retrying.** Sync the lists, then retry. |
| 500 | `{"state":"error",…}` | Nothing was saved, because the save is all-or-nothing. Retry with backoff. |

The package save (`POST /manifest/inspection/save/package/inspection`) and the remarks-only save (`POST /manifest/inspection/save/remarks/only`) use the same 404 and 422 codes.

### Uploading files

`POST /manifest/media/upload/file` (multipart, field `file`):

| HTTP | Meaning | What the app does |
|---|---|---|
| 200 | `data` is the stored path, for example `images/abc_1791299974.jpg`. Sending the same content again returns the path of the copy already stored, so retries no longer fill the disk. | Store the path; never upload that file again. |
| 400 | No file in the request. | Stop; it is a bug in the request. |
| 408 | The upload stopped arriving (weak Wi-Fi, dropped connection). | Retry with backoff. |
| 422 | The file is empty. | Stop; ask the user to replace the file. |

An upload may take up to 30 minutes, so large videos on slow Wi-Fi are not cut off. Other request bodies must arrive within 2 minutes.

### Queue rules

- **One queue worker at a time.** Use a mutex or a single isolate. The connectivity listener and the timer must not flush the queue in parallel. The server logs show bursts of the same save sent several times within one second.
- **Upload each photo once.** Upload it through `/manifest/media/upload/file`, store the returned path in the queue item, and reuse that path on retries. Today every retry uploads the photo again, and the server keeps a new copy each time.
- **Back off between retries:** 5 s, 10 s, 20 s, 40 s, then every 5 minutes. Do not retry every 7 s forever.
- **Send real paths.** `vehicleImage` and the other media fields must hold the stored path, for example `images/abc.jpg`. Never send the `toString()` of a Dart map. Older logs show values like `{filename: …, path: …}`.

## 2. Knowing what other tablets inspected

### Status endpoint

`GET /manifest/vehicles/status/:manifestId?since=<cursor>`

```json
{
  "state": "success",
  "full": false,
  "cursor": "2026-10-06 14:02:30.481000",
  "data": [
    {
      "vehicleId": "…",
      "chasisNumber": "…",
      "blNumber": "…",
      "inspectionStatus": "yes",
      "isInspected": true,
      "inspectionTime": "2026-10-06 09:00:00",
      "inspectedBy": "Asha One",
      "isPublished": "no",
      "hasHistory": false,
      "updatedAt": "2026-10-06 14:02:29.636000"
    }
  ]
}
```

- **Without `since`:** the response holds every vehicle of the manifest, and `full` is `true`.
- **With `since`:** the response holds only the vehicles that changed since the cursor.
- **Cursor:** store `cursor` and send it as `since` on the next call. The cursor lies a few seconds in the past, so a row can come back twice. Merge rows by `vehicleId`, and the newest `updatedAt` wins.
- **`full: true` with a cursor:** this means the database migration has not been applied yet. Replace the local list with the response.

### Live events

`GET /sync/events/:manifestId` is a Server-Sent Events stream.

| Event | Data | What the app does |
|---|---|---|
| `ready` | `{"manifestId": "…"}` | Call the status endpoint. This covers anything missed while disconnected. |
| `vehicle` | `{"vehicleId": "…", "kind": "inspected" \| "reinspected" \| "remarks" \| "published", "at": "…"}` | Call the status endpoint with the stored cursor. |
| (comment) | `: keep-alive`, every 15 s | Nothing. If none arrives for 45 s, reconnect. |

Events are only a trigger. Always read the actual data from the status endpoint.

### When to call the status endpoint

1. When the manifest opens: call it without `since`, and store the result in the local DB.
2. On every `ready` or `vehicle` event.
3. Every 15 s while the manifest screen is open. This is the fallback when the event stream is down.
4. When the app resumes, and when the network comes back.

### Before inspecting a chassis

Call `GET /manifest/vehicles/list/inspection/check/:manifestId?query=<chassis>`. If `isInspected` is true, show who inspected the vehicle and when. Ask the user before sending `reinspect: true`.

## 3. Comparing a tablet with the server

`POST /manifest/vehicles/compare/:manifestId` compares what the tablet holds with what the server holds. The server also stores the report, so an operator can see which tablet still has unpublished or mismatched work.

Request:

```json
{
  "device": {"deviceId": "…", "deviceName": "Tablet 7 (Asha)", "appVersion": "2.1.0"},
  "vehicles": [
    {"vehicleId": "…", "chasisNumber": "…", "inspected": true, "inspectionTime": "2026-10-06 09:00:00",
     "submissionId": "…", "published": true, "queueState": "",
     "checks": [["checkId", "seen/okay"]], "remarks": ["side lights"], "mediaCount": 3, "onboardPackageCount": 1}
  ],
  "packages": [
    {"packageId": "…", "packageNumber": "…", "inspected": true, "published": true,
     "typeId": "…", "inspectionStatusId": "…", "inspectionTime": "…"}
  ]
}
```

- `inspected` is true only for the tablet's own inspections. A vehicle downloaded with another tablet's inspection counts as not inspected.
- `inspectionTime` is the time the tablet sends as `inspectionTime` when publishing.

The response lists only the items that differ. Each item has a status and the action the tablet offers:

| Status | Meaning | Action |
|---|---|---|
| `missing_on_server` | The tablet inspected it; the server has no inspection yet. | `publish` |
| `published_but_missing` | The tablet marked it published, but the server has no inspection. | `republish` |
| `published_not_marked` | The server already has this exact inspection; the tablet has not marked it published. | `mark_published` |
| `inspected_elsewhere` | The server has an inspection; this tablet has none. | `pull_status` |
| `content_mismatch` | Same inspection, but check results or remarks differ. | `republish` |
| `different_inspection` | Both have an inspection, but not the same one (another inspector or another time). | `review` |
| `unknown_on_server` | The server does not know the vehicle or package. | `review` |
| `missing_on_tablet` | The server has a vehicle the tablet does not. | `refresh_manifest` |

```json
{"state": "success", "data": {
  "reportId": "…", "comparedAt": "…", "stored": true,
  "summary": {"vehicles": {"published_but_missing": 1, "missing_on_tablet": 20}, "packages": {}},
  "items": [{"kind": "vehicle", "id": "…", "label": "CH-01", "status": "published_but_missing", "action": "republish",
             "differences": ["…"], "info": ["…"], "server": {"inspected": false, "inspectedBy": "", "…": "…"}}]
}}
```

Media count differences appear under `info` only; they never change the status.

Stored reports:

- `GET /manifest/vehicles/compare/reports/:manifestId`: the latest report of every tablet, with its summary and `needsAction` count.
- `GET /manifest/vehicles/compare/report/:reportId`: one report with all its items.

## 4. Publishing

`GET /sync/trigger/publish/:manifestId` and `GET /sync/trigger/added-later/publish/:manifestId` work as before. Several things changed:

- **The job runs on the server.** If the tablet disconnects, the job keeps running. If a second tablet triggers a publish while one runs, it attaches to the running job and gets its progress. The upload does not run twice.
- **There is a new event, `warning`.** It is sent when a non-essential file is missing, for example a fault photo. The vehicle is still published, without that file.
- **The `complete` event reports the outcome:**

  ```json
  {
    "state": "success | partial | error",
    "data": {
      "message": "…",
      "total": 10,
      "published": 8,
      "failed": 2,
      "failedItems": [
        {"kind": "vehicle", "id": "…", "label": "<chassis>", "reason": "image file missing on server (images/x.jpg)"}
      ],
      "warnings": ["…"],
      "progress": 100,
      "status": "finished"
    }
  }
  ```

  Show `failedItems` to the user. `state: "partial"` means some items failed. Do not show "published" for those items.
- **The vehicle list includes the publish status.** `GET /manifest/vehicles/list/local/:manifestId` now returns `isPublished`, `inspectedBy` and `inspectionCount` for each vehicle. Read `isPublished` from there, or from the status endpoint. Do not infer it from the publish stream.
