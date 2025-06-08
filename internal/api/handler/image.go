package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"skidimg/internal/model"
	"skidimg/internal/token"
	"strings"

	"github.com/go-chi/chi"
	"github.com/h2non/bimg"
)

func min(num1, num2 int) int {
	if num1 < num2 {
		return num1
	}
	return num2
}
func generateShortID(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

func (h *handler) renderUploadPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/upload.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)
	tmpl.ExecuteTemplate(w, "layout", layoutData)
}

func (h *handler) uploadImage(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".png"
	}

	id := generateShortID(8)
	origPath := fmt.Sprintf("uploads/original/%s%s", id, ext)
	optPath := fmt.Sprintf("uploads/optimized/%s.webp", id)

	origData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read image data", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(origPath, origData, 0644); err != nil {
		http.Error(w, "Failed to save original image", http.StatusInternalServerError)
		return
	}

	size, _ := bimg.Size(origData)
	imgWidth := min(size.Width, 1280)
	newImage, err := bimg.NewImage(origData).Process(bimg.Options{
		Width:   imgWidth,
		Quality: 80,
		Type:    bimg.WEBP,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to optimize image %v", err), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(optPath, newImage, 0644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save optimized image %v", err), http.StatusInternalServerError)
		return
	}

	claimsVal := r.Context().Value(authKey{})
	var userID *int64

	if claimsVal != nil {
		if claims, ok := claimsVal.(*token.UserClaims); ok {
			userID = &claims.ID
		}
	}

	img := &model.Image{
		UserID:   userID,
		Filename: id,
		Ext:      ext,
	}
	_, err = h.server.CreateImage(h.ctx, img)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save image metadata %v", err), http.StatusInternalServerError)
		return
	}

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(model.ImageRes{
	// 	ID:       img.ID,
	// 	Filename: id,
	// })

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "/gallery/%s", id)
}

func (h *handler) viewImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	img, err := h.server.GetImageByFilename(h.ctx, id)
	if err != nil {
		log.Printf("failed to get image by filename: %v", err)
		http.NotFound(w, r)
		return
	}

	ua := r.UserAgent()
	lowerUA := strings.ToLower(ua)

	isBot := strings.Contains(lowerUA, "telegrambot") ||
		strings.Contains(lowerUA, "whatsapp") ||
		strings.Contains(lowerUA, "discord") ||
		strings.Contains(lowerUA, "facebookexternal") ||
		strings.Contains(lowerUA, "twitterbot") ||
		strings.Contains(lowerUA, "slack") ||
		strings.Contains(lowerUA, "preview")

	if isBot {
		// Отдаём оптимизированную webp напрямую
		http.ServeFile(w, r, fmt.Sprintf("uploads/optimized/%s.webp", img.Filename))
		return
	}

	// Это браузер — рендерим HTML
	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/view.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	origPath := fmt.Sprintf("/uploads/original/%s%s", img.Filename, img.Ext)
	data := struct {
		ImagePath string
	}{
		ImagePath: origPath,
	}

	tmpl.ExecuteTemplate(w, "layout", data)

}

func (h *handler) handleGalleryPage(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	images, err := h.server.GetUserGalary(h.ctx, claims.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading gallery %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/gallery.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}
	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)
	layoutData.Content = images

	err = tmpl.ExecuteTemplate(w, "layout", layoutData)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

func (h *handler) renderProfilePage(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/profile.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)
	tmpl.ExecuteTemplate(w, "layout", layoutData)

}
