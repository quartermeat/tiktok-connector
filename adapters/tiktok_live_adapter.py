"""Forward real TikTok LIVE events into the local normalized connector.

This uses the unofficial TikTokLive Webcast client. It does not need a
TikTok password or stream key; set TIKTOK_USERNAME to the public creator
handle that is currently LIVE.
"""

import json
import os
import sys
import urllib.request

from TikTokLive import TikTokLiveClient
from TikTokLive.events import CommentEvent, ConnectEvent, FollowEvent, GiftEvent, JoinEvent, LikeEvent, LiveEndEvent

CONNECTOR_POST_URL = os.environ.get("CONNECTOR_POST_URL", "http://127.0.0.1:8787/api/events")
USERNAME = os.environ.get("TIKTOK_USERNAME", "").lstrip("@").strip()

if not USERNAME:
    raise SystemExit("Set TIKTOK_USERNAME first, for example: $env:TIKTOK_USERNAME='your_handle'")

client = TikTokLiveClient(unique_id=USERNAME)

def avatar_url(user):
    for candidate in (
        getattr(user, "avatar_thumb", None),
        getattr(user, "avatar_medium", None),
        getattr(user, "avatar_larger", None),
    ):
        urls = getattr(candidate, "url_list", None) or getattr(candidate, "urls", None)
        if urls:
            return urls[0]
    return None

def user_data(user):
    return {
        "user_id": str(getattr(user, "user_id", "") or getattr(user, "unique_id", "")),
        "avatar_url": avatar_url(user),
    }

def send_event(event_type, user, **fields):
    payload = {
        "source": "tiktok-live",
        "type": event_type,
        "user": getattr(user, "unique_id", None) or getattr(user, "nickname", None) or "viewer",
        "raw": user_data(user),
        **fields,
    }
    request = urllib.request.Request(
        CONNECTOR_POST_URL,
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=10) as response:
        if response.status >= 300:
            raise RuntimeError(f"connector returned HTTP {response.status}")
    print(json.dumps(payload, ensure_ascii=False), flush=True)

@client.on(ConnectEvent)
async def on_connect(event: ConnectEvent):
    print(f"Connected to @{USERNAME}; room={client.room_id}", flush=True)

@client.on(CommentEvent)
async def on_comment(event: CommentEvent):
    send_event("comment", event.user, text=event.comment)

@client.on(GiftEvent)
async def on_gift(event: GiftEvent):
    gift = getattr(event.gift, "name", None) or "Gift"
    repeat_count = int(getattr(event, "repeat_count", 1) or 1)
    send_event("gift", event.user, gift=gift, value=repeat_count)

@client.on(LikeEvent)
async def on_like(event: LikeEvent):
    send_event("like", event.user, value=int(getattr(event, "count", 1) or 1))

@client.on(FollowEvent)
async def on_follow(event: FollowEvent):
    send_event("follow", event.user)

@client.on(JoinEvent)
async def on_join(event: JoinEvent):
    send_event("join", event.user)

@client.on(LiveEndEvent)
async def on_live_end(event: LiveEndEvent):
    send_event("live_end", event.user if getattr(event, "user", None) else type("LiveHost", (), {"unique_id": USERNAME, "nickname": USERNAME})())

if __name__ == "__main__":
    try:
        client.run()
    except KeyboardInterrupt:
        print("Adapter stopped.", file=sys.stderr)
