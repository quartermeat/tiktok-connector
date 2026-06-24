package connector

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	hub *Hub
}

func NewServer() *Server {
	return &Server{hub: NewHub()}
}

func (s *Server) Hub() *Hub {
	return s.hub
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/events", s.handleEvents)
	mux.HandleFunc("/api/events", s.handleAPIEvents)
	mux.HandleFunc("/api/events/", s.handleAPIEvent)
	mux.HandleFunc("/api/health", s.handleHealth)
	return withCORS(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"recent": len(s.hub.Recent()),
	})
}

func (s *Server) handleAPIEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.hub.Recent())
	case http.MethodPost:
		var event StreamEvent
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			event = StreamEvent{
				Source: r.FormValue("source"),
				Type:   r.FormValue("type"),
				User:   r.FormValue("user"),
				Text:   r.FormValue("text"),
				Gift:   r.FormValue("gift"),
			}
			if value := r.FormValue("value"); value != "" {
				if parsed, err := strconv.Atoi(value); err == nil {
					event.Value = parsed
				}
			}
		}
		writeJSON(w, http.StatusAccepted, s.hub.Publish(event))
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAPIEvent(w http.ResponseWriter, r *http.Request) {
	idText := strings.TrimPrefix(r.URL.Path, "/api/events/")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid event id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if !s.hub.Consume(id) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")

	for _, event := range s.hub.Recent() {
		writeSSE(w, event)
	}
	flusher.Flush()

	ch, cancel := s.hub.Subscribe()
	defer cancel()
	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-ch:
			writeSSE(w, event)
			flusher.Flush()
		}
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = indexTemplate.Execute(w, nil)
}

func writeSSE(w http.ResponseWriter, event StreamEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Type, data)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Private-Network", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>TikTok Connector</title>
  <style>
    :root { color-scheme: dark; --bg:#080d12; --panel:#101922; --line:#263746; --text:#edf6f2; --muted:#9eb3ba; --accent:#67e8c2; }
    * { box-sizing: border-box; }
    body { margin: 0; min-height: 100vh; background: #080d12; color: var(--text); font-family: ui-sans-serif, system-ui, Segoe UI, sans-serif; }
    main { width: min(100vw, 980px); margin: 0 auto; padding: 20px; }
    header, form, section { border-bottom: 1px solid var(--line); padding: 14px 0; }
    h1 { margin: 0; font-size: 20px; }
    p { color: var(--muted); }
    a { color: var(--accent); }
    .actions { display: flex; gap: 10px; flex-wrap: wrap; align-items: center; }
    .app-link { display: inline-flex; align-items: center; height: 36px; padding: 0 12px; color: #052018; background: var(--accent); text-decoration: none; font-weight: 700; }
    form { display: grid; grid-template-columns: 140px 120px 1fr 90px auto; gap: 8px; align-items: end; }
    label { display: grid; gap: 5px; color: var(--muted); font-size: 12px; }
    input, select, button { height: 36px; border: 1px solid var(--line); background: var(--panel); color: var(--text); padding: 0 10px; }
    button { color: #052018; background: var(--accent); border: 0; font-weight: 700; cursor: pointer; }
    pre { min-height: 280px; margin: 0; padding: 12px; overflow: auto; background: #05090d; border: 1px solid var(--line); color: var(--text); }
    code { color: var(--accent); }
  </style>
</head>
<body>
  <main>
    <header>
      <h1>TikTok Connector</h1>
      <p>Local normalized event bridge for the hosted Wellfield app. Keep this service running on <code>127.0.0.1:8787</code>.</p>
      <div class="actions">
        <a class="app-link" href="https://quartermeat.github.io/tiktok-connector/" target="_blank" rel="noreferrer">Open hosted app</a>
        <a href="https://quartermeat.github.io/tiktok-connector/remote/" target="_blank" rel="noreferrer">Open remote connector</a>
        <span>Test events post to <code>/api/events</code>.</span>
      </div>
    </header>
    <form id="event-form">
      <label>Type
        <select name="type">
          <option value="comment">comment</option>
          <option value="gift">gift</option>
          <option value="like">like</option>
          <option value="follow">follow</option>
        </select>
      </label>
      <label>User
        <input name="user" value="viewer">
      </label>
      <label>Text / Gift
        <input name="text" value="!attract">
      </label>
      <label>Value
        <input name="value" type="number" value="1">
      </label>
      <button type="submit">Send</button>
    </form>
    <section>
    <pre id="log"></pre>
    </section>
  </main>
  <script>
    const log = document.getElementById("log");
    const form = document.getElementById("event-form");
    async function refreshRecent() {
      const res = await fetch("/api/events");
      const events = await res.json();
      log.textContent = events.length ? JSON.stringify(events, null, 2) : "No pending events.";
    }
    const stream = new EventSource("/events");
    ["comment", "gift", "like", "follow"].forEach((type) => {
      stream.addEventListener(type, refreshRecent);
    });
    setInterval(refreshRecent, 1000);
    refreshRecent();
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const data = Object.fromEntries(new FormData(form).entries());
      if (data.type === "gift") {
        data.gift = data.text;
        data.text = "";
      }
      data.source = "manual";
      data.value = Number(data.value || 1);
      const res = await fetch("/api/events", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data)
      });
      await res.json();
      refreshRecent();
    });
  </script>
</body>
</html>`))
