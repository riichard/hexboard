package main

import (
	"html/template"
	"net/http"
	"sync"
	"time"

	"post6.net/gohexdump/internal/screen"
)

const maxRecent = 10

type webHandler struct {
	screenChan chan<- screen.Screen
	timeout    time.Duration
	mu         sync.Mutex
	recent     []string
}

func (h *webHandler) send(msg string) {
	h.mu.Lock()
	h.recent = append([]string{msg}, h.recent...)
	if len(h.recent) > maxRecent {
		h.recent = h.recent[:maxRecent]
	}
	h.mu.Unlock()

	h.screenChan <- newMessageScreen(msg)
	go func() {
		time.Sleep(h.timeout)
		h.screenChan <- newRainScreen()
	}()
}

func (h *webHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if msg := r.FormValue("message"); msg != "" {
			h.send(msg)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	h.mu.Lock()
	recent := append([]string{}, h.recent...)
	h.mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexTmpl.Execute(w, recent)
}

func startWebServer(addr string, screenChan chan<- screen.Screen, timeout time.Duration) {
	h := &webHandler{screenChan: screenChan, timeout: timeout}
	http.ListenAndServe(addr, h)
}

var indexTmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Hexboard</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    background: #0d0d0d;
    color: #00ff41;
    font-family: 'Courier New', Courier, monospace;
    min-height: 100dvh;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2rem 1.25rem 3rem;
    gap: 2rem;
  }

  header {
    text-align: center;
    letter-spacing: 0.25em;
    font-size: 0.8rem;
    opacity: 0.5;
    text-transform: uppercase;
  }

  h1 {
    font-size: clamp(1.4rem, 6vw, 2rem);
    letter-spacing: 0.3em;
    font-weight: normal;
  }

  form.compose {
    width: 100%;
    max-width: 520px;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  input[type="text"] {
    width: 100%;
    padding: 1rem 1.1rem;
    font-size: clamp(1rem, 4vw, 1.25rem);
    font-family: inherit;
    background: #111;
    color: #00ff41;
    border: 1px solid #00ff41;
    border-radius: 6px;
    outline: none;
    -webkit-appearance: none;
  }
  input[type="text"]::placeholder { color: #1a4d2a; }
  input[type="text"]:focus {
    border-color: #7fff7f;
    box-shadow: 0 0 0 3px #00ff4122;
  }

  button.send {
    width: 100%;
    padding: 1rem;
    font-size: clamp(1rem, 4vw, 1.2rem);
    font-family: inherit;
    font-weight: bold;
    letter-spacing: 0.2em;
    background: #00ff41;
    color: #0d0d0d;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    -webkit-tap-highlight-color: transparent;
    transition: background 0.1s;
  }
  button.send:active { background: #7fff7f; }

  .recent {
    width: 100%;
    max-width: 520px;
  }

  .recent-label {
    font-size: 0.7rem;
    letter-spacing: 0.2em;
    opacity: 0.4;
    margin-bottom: 0.75rem;
  }

  form.recent-item {
    border-bottom: 1px solid #1a1a1a;
  }

  button.recent-btn {
    width: 100%;
    padding: 0.75rem 0.25rem;
    font-family: inherit;
    font-size: 0.95rem;
    background: none;
    color: #00cc33;
    border: none;
    text-align: left;
    cursor: pointer;
    opacity: 0.7;
    -webkit-tap-highlight-color: transparent;
  }
  button.recent-btn::before { content: '> '; opacity: 0.4; }
  button.recent-btn:active { opacity: 1; color: #00ff41; }
</style>
</head>
<body>
  <header>
    <h1>[ HEXBOARD ]</h1>
  </header>

  <form class="compose" method="POST" action="/">
    <input type="text" name="message" placeholder="type a message..."
           autofocus autocomplete="off" autocorrect="off" autocapitalize="off"
           spellcheck="false" maxlength="128">
    <button class="send" type="submit">SEND</button>
  </form>

  {{if .}}
  <div class="recent">
    <div class="recent-label">RECENT</div>
    {{range .}}
    <form class="recent-item" method="POST" action="/">
      <input type="hidden" name="message" value="{{.}}">
      <button class="recent-btn" type="submit">{{.}}</button>
    </form>
    {{end}}
  </div>
  {{end}}
</body>
</html>
`))
