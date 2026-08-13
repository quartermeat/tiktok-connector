# TikTok Connector

Local stream interaction bridge for Codex-built apps.

The connector normalizes incoming stream activity into a simple event schema and exposes it over Server-Sent Events. The first version includes a manual test console so apps can integrate before a real TikTok LIVE adapter is plugged in.

The repo also hosts the Wellfield WASM app through GitHub Pages from `docs/`.

## Run the Local Connector

```powershell
cd D:\Codex\Projects\work\tiktok-connector
.\scripts\run.ps1
```

Open <http://127.0.0.1:8787> for the local test console.

## Hosted Page

The GitHub Pages site is the default Wellfield app:

<https://quartermeat.github.io/tiktok-connector/>

The hosted page can still consume stream events from the local connector when this service is running on `127.0.0.1:8787`.

To open the hosted app from PowerShell:

```powershell
.\scripts\open-app.ps1
```

## Remote Connector

The public remote connector page is:

<https://quartermeat.github.io/tiktok-connector/remote/>

It publishes viewer commands into the shared relay topic `quartermeat-tiktok-connector`. The local connector subscribes to that topic by default and republishes incoming remote events to the game.

Remote relay flags:

```powershell
.\scripts\run.ps1
go run ./cmd/tiktok-connector -remote-topic quartermeat-tiktok-connector
go run ./cmd/tiktok-connector -remote-topic ""
```

The empty topic disables remote relay ingestion.

## Real TikTok LIVE adapter

The optional adapter in `adapters/tiktok_live_adapter.py` reads real comments, gifts, likes, and follows from an active LIVE room and posts normalized events to this connector. It uses the unofficial `TikTokLive` Webcast client and requires only the public creator username; never provide a TikTok password, cookie, session token, or stream key.

Install and run it in a separate PowerShell window:

```powershell
cd D:\Codex\Projects\work\tiktok-connector
python -m pip install -r .\adapters\requirements.txt
$env:TIKTOK_USERNAME='your_public_handle'
.\scripts\run-tiktok-adapter.ps1
```

Start the connector first, then start the adapter only after the creator is LIVE. The adapter is experimental and unofficial, so verify it with a short test LIVE before relying on it for a public broadcast.

Refresh the hosted bundle from the local Wellfield build with:

```powershell
.\scripts\sync-site.ps1
```

## Endpoints

- `GET /` test console
- `GET /events` Server-Sent Events stream
- `GET /api/events` recent replay buffer
- `POST /api/events` inject a normalized event
- `DELETE /api/events/{id}` mark an event as consumed and remove it from the replay buffer
- `GET /api/health` health check

## Event Shape

```json
{
  "id": 1,
  "source": "manual",
  "type": "comment",
  "user": "viewer_name",
  "text": "!attract",
  "value": 1,
  "command": "attract",
  "args": [],
  "receivedAt": "2026-06-24T00:00:00Z"
}
```

## App Integration

Browser/WASM apps can subscribe directly:

```js
const stream = new EventSource("http://127.0.0.1:8787/events");
stream.addEventListener("comment", (event) => {
  const payload = JSON.parse(event.data);
  console.log(payload.command, payload.user);
});
```

After an app consumes an event, it should acknowledge the event so it does not stay in the pending list or replay on reconnect:

```js
await fetch(`http://127.0.0.1:8787/api/events/${payload.id}`, { method: "DELETE" });
```

## TikTok Adapter

This project intentionally starts with the local normalized bridge. A real TikTok LIVE adapter should feed events into `POST /api/events` or call `Hub.Publish` directly once the capture approach is chosen.
