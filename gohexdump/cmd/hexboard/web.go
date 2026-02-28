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

  h1 {
    font-size: clamp(1.3rem, 6vw, 1.8rem);
    letter-spacing: 0.3em;
    font-weight: normal;
    opacity: 0.6;
  }

  form.compose {
    width: 100%;
    max-width: 520px;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .field-wrap {
    position: relative;
  }

  textarea {
    width: 100%;
    padding: 0.85rem 1rem;
    font-size: clamp(0.95rem, 3.5vw, 1.1rem);
    font-family: inherit;
    background: #111;
    color: #00ff41;
    border: 1px solid #00ff41;
    border-radius: 6px;
    outline: none;
    resize: none;
    line-height: 1.6;
    /* fixed height: 4 rows */
    height: calc(4 * 1.6em + 1.7rem);
    -webkit-appearance: none;
    overflow: hidden;
  }
  textarea::placeholder { color: #1a4d2a; }
  textarea:focus {
    border-color: #7fff7f;
    box-shadow: 0 0 0 3px #00ff4122;
  }

  /* row ruler — 4 faint dividers behind the textarea */
  .row-guides {
    position: absolute;
    inset: 1px;
    border-radius: 5px;
    pointer-events: none;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
  .row-guide {
    flex: 1;
    border-bottom: 1px solid #1c1c1c;
  }
  .row-guide:last-child { border-bottom: none; }

  .hint {
    font-size: 0.68rem;
    opacity: 0.35;
    letter-spacing: 0.05em;
    text-align: right;
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
    margin-top: 0.15rem;
  }
  button.send:active { background: #7fff7f; }

  .recent {
    width: 100%;
    max-width: 520px;
  }

  .recent-label {
    font-size: 0.7rem;
    letter-spacing: 0.2em;
    opacity: 0.35;
    margin-bottom: 0.6rem;
  }

  form.recent-item {
    border-bottom: 1px solid #181818;
  }

  button.recent-btn {
    width: 100%;
    padding: 0.65rem 0.25rem;
    font-family: inherit;
    font-size: 0.9rem;
    background: none;
    color: #00cc33;
    border: none;
    text-align: left;
    cursor: pointer;
    opacity: 0.65;
    white-space: pre;
    overflow: hidden;
    text-overflow: ellipsis;
    -webkit-tap-highlight-color: transparent;
  }
  button.recent-btn::before { content: '> '; opacity: 0.4; }
  button.recent-btn:active { opacity: 1; color: #00ff41; }
</style>
</head>
<body>
  <h1>[ HEXBOARD ]</h1>

  <form class="compose" method="POST" action="/">
    <div class="field-wrap">
      <div class="row-guides">
        <div class="row-guide"></div>
        <div class="row-guide"></div>
        <div class="row-guide"></div>
        <div class="row-guide"></div>
      </div>
      <textarea name="message" rows="4"
                placeholder="line 1&#10;line 2&#10;line 3&#10;line 4"
                autofocus autocomplete="off" autocorrect="off"
                autocapitalize="off" spellcheck="false"></textarea>
    </div>
    <div class="hint">4 rows &times; 32 chars</div>
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
