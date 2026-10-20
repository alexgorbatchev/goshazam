---
created_on: 2026-10-19 17:40
last_modified: 2026-10-19 17:40
status: current
---

# Tests

Run `go test ./...` for the local test suite. Shazam API responses use local HTTP servers; the default suite makes no live Shazam requests.

Live tests require the `integration` build tag:

```sh
go test -tags integration -run '^TestShazamLive' -count=1 -v .
```

The recognition test requires ffmpeg and `SHAZAM_TEST_AUDIO` pointing to ShazamIO's `examples/data/Gloria.ogg`; it skips when either prerequisite is absent. The related-tracks test runs without an audio fixture. Both tests skip upstream rate limiting (HTTP 429). The Live Shazam integration workflow supplies the pinned, checksum-verified fixture and ffmpeg, and runs manually or weekly.

Shazam can return HTTP 204 for related tracks. `RelatedTracks` returns `{"tracks":[]}` without an error for that response; an empty list does not demonstrate that upstream recommendations are available. The shared JSON client reports `models.ErrNoContent` for HTTP 204 so callers can choose their endpoint-specific behavior. Empty or malformed HTTP 200 bodies remain JSON errors, and HTTP failures remain `models.APIError`.
