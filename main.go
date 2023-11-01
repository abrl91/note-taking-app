package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

var notes []Note
var id int // 0 by default
var mu sync.Mutex

func main() {
	http.HandleFunc("/notes", notesHandler)
	http.ListenAndServe(":8080", nil)
}

func notesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getNotes(w, r)
	case http.MethodPost:
		addNotes(w, r)
	case http.MethodPut:
		updateNotes(w, r)
	case http.MethodDelete:
		deleteNotes(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
	}
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	// check if an id is provided as a query parameter
	ids, ok := r.URL.Query()["id"]

	// if id is provided, retrieve that specific note
	if ok && len(ids) == 1 {
		id, err := strconv.Atoi(ids[0])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid ID format"))
			return
		}

		for _, note := range notes {
			if note.ID == id {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(note)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Note not found"))
		return
	}

	// if no id is provided - retrieve all notes
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func addNotes(w http.ResponseWriter, r *http.Request) {
	var note Note

	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return
	}

	mu.Lock()
	defer mu.Unlock()

	id++
	note.ID = id
	notes = append(notes, note)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func updateNotes(w http.ResponseWriter, r *http.Request) {
	var updateNote Note

	err := json.NewDecoder(r.Body).Decode(&updateNote)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to decode note"))
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i, note := range notes {
		if note.ID == updateNote.ID {
			notes[i] = updateNote
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(updateNote)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Note Not Found"))

}

func deleteNotes(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query()["id"]

	if len(ids) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID is required"))
		return
	}

	// convert the id string from the query param to an int
	id, err := strconv.Atoi(ids[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid ID"))
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i, note := range notes {
		if note.ID == id {
			// notes[:i] is everything before the note we want to delete
			// notes[i+1:] is everything after the note we want to delete
			// append combines the two slices
			notes = append(notes[:i], notes[i+1:]...)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Note Not Found"))
}
