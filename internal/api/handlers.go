package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"secret-share/internal/crypto"
	"secret-share/internal/models"
	"secret-share/internal/store"
)

type Server struct {
	store store.Store
}

func NewServer(s store.Store) *Server {
	return &Server{store: s}
}

// Responses
type CreateSecretResponse struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type RetrieveSecretRequest struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

func (s *Server) HandleCreateSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload models.SecretPayload

	// Check Content-Type to decide how to parse
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Limit total request size (e.g. 100MB file + overhead)
		// We'll set a hard limit of 110MB for safety.
		r.Body = http.MaxBytesReader(w, r.Body, 110*1024*1024)
		if err := r.ParseMultipartForm(110 * 1024 * 1024); err != nil {
			http.Error(w, "File too large or invalid multipart", http.StatusBadRequest)
			return
		}

		// Check for file
		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			buf := make([]byte, header.Size)
			if _, err := io.ReadFull(file, buf); err != nil {
				http.Error(w, "Failed to read file", http.StatusInternalServerError)
				return
			}
			payload.IsFile = true
			payload.FileData = buf
			payload.Filename = header.Filename
			payload.MimeType = header.Header.Get("Content-Type")
		} else {
			// No file, maybe just text field?
			text := r.FormValue("data")
			if text == "" {
				http.Error(w, "No secret data provided", http.StatusBadRequest)
				return
			}
			payload.Text = text
		}

	} else {
		// JSON
		r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) // 1MB limit for JSON text
		var req struct {
			Data string `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.Data == "" {
			http.Error(w, "Data cannot be empty", http.StatusBadRequest)
			return
		}
		payload.Text = req.Data
	}

	// Serialize Payload to JSON for encryption
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Encoding error", http.StatusInternalServerError)
		return
	}

	// 1. Generate ID (12 bytes -> 16 chars Base64URL)
	idBytes := make([]byte, 12)
	if _, err := rand.Read(idBytes); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	id := base64.RawURLEncoding.EncodeToString(idBytes)

	// 2. Generate Key (12 bytes -> 16 chars Base64URL)
	keyBytes := make([]byte, 12)
	if _, err := rand.Read(keyBytes); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	keyStr := base64.RawURLEncoding.EncodeToString(keyBytes)

	// 3. Expand Key to 32 bytes using SHA256 for AES-256
	// This allows us to use a shorter key in the URL
	encryptionKey := sha256.Sum256([]byte(keyStr))

	// 4. Encrypt using expanded key
	encrypted, err := crypto.Encrypt(payloadBytes, encryptionKey[:])
	if err != nil {
		http.Error(w, "Encryption failed", http.StatusInternalServerError)
		return
	}

	// 5. Save to Store
	if err := s.store.Save(id, encrypted); err != nil {
		// Log the actual error for debugging
		http.Error(w, "Failed to store secret: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Respond with ID and Key
	resp := CreateSecretResponse{
		ID:  id,
		Key: keyStr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleRetrieveSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RetrieveSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id := strings.TrimSpace(req.ID)
	keyStr := strings.TrimSpace(req.Key)

	if id == "" || keyStr == "" {
		http.Error(w, "ID and Key are required", http.StatusBadRequest)
		return
	}

	// 1. Retrieve
	encryptedData, err := s.store.Get(id)
	if err != nil {
		if err == store.ErrBurned {
			http.Error(w, "Secret already retrieved (burned)", http.StatusGone) // 410
			return
		}
		if err == store.ErrRecycled {
			http.Error(w, "Secret evicted to free up memory (recycled)", http.StatusGone) // 410
			return
		}
		if err == store.ErrNotFound {
			http.Error(w, "Secret not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 2. KEY EXPANSION: SHA256 hash the incoming short key to get 32 bytes
	decryptionKey := sha256.Sum256([]byte(keyStr))

	// 3. Decrypt using expanded key
	decrypted, err := crypto.Decrypt(encryptedData, decryptionKey[:])
	if err != nil {
		http.Error(w, "Decryption failed (invalid key)", http.StatusUnauthorized)
		return
	}

	// 3. Unmarshal Payload
	// For backward compatibility (if any old secrets exist), check logic?
	// But in-memory store is wiped on restart, so no worries.
	var payload models.SecretPayload
	if err := json.Unmarshal(decrypted, &payload); err != nil {
		// Fallback: It might be old format (raw string)?
		// Actually, we just started, so let's assume it works.
		// If it fails, maybe it IS just text?
		payload.Text = string(decrypted)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}
