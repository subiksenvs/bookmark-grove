# Bookmark Grove

A personal knowledge library for collecting useful links with context and tags.

## Features
- Go standard-library HTTP backend and JSON API.
- Persistent local JSON library with create, search, tagging, and delete.
- Optional OpenAI-powered organization coach; demo answer works without a key.
- Responsive UI with safe external-link handling.

## Run
Requires Go 1.23+. Run `go run .` and open http://127.0.0.1:8792. Set `PORT` or `OPENAI_API_KEY` in your environment to configure the app. Data is stored in `data/library.json`.

## API
`GET /api/links?q=...`, `POST /api/links`, `DELETE /api/links/:id`, `GET /api/summary`, `POST /api/coach`.
