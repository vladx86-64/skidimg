package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"skidimg/internal/model"
	"skidimg/internal/token"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/h2non/bimg"
)

func (h *handler) renderAlbumPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/albums.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)

	tmpl.ExecuteTemplate(w, "layout", layoutData)
}

func (h *handler) renderUserAlbums(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	albums, err := h.server.GetUserAlbums(h.ctx, claims.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load albums: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/albums.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)
	layoutData.Content = albums
	layoutData.Title = "Your Albums"

	tmpl.ExecuteTemplate(w, "layout", layoutData)
}

func (h *handler) handleCreateAlbum(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	file, _, err := r.FormFile("preview")
	if err != nil {
		http.Error(w, "Preview file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	origData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	// Пресет: фиксированная ширина, качественная обрезка
	previewData, err := bimg.NewImage(origData).Process(bimg.Options{
		Width:   400,
		Height:  250,
		Crop:    true,
		Quality: 80,
		Type:    bimg.WEBP,
	})
	if err != nil {
		http.Error(w, "Image conversion error", http.StatusInternalServerError)
		return
	}

	album := &model.Album{
		UserID:      claims.ID,
		Title:       title,
		Description: description,
	}

	savedAlbum, err := h.server.CreateAlbum(h.ctx, album)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating album: %v", err), http.StatusInternalServerError)
		return
	}

	// Сохраняем превью
	previewPath := fmt.Sprintf("uploads/albums/preview/%d.webp", savedAlbum.ID)
	if err := os.WriteFile(previewPath, previewData, 0644); err != nil {
		http.Error(w, "Error saving preview", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/albums", http.StatusSeeOther)
}

func (h *handler) viewAlbum(w http.ResponseWriter, r *http.Request) {
	albumIDStr := chi.URLParam(r, "id")
	albumID, _ := strconv.ParseInt(albumIDStr, 10, 64)

	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	albumImages, err := h.server.GetImagesByAlbum(h.ctx, albumID)
	if err != nil {
		http.Error(w, "Failed to load album images", http.StatusInternalServerError)
		return
	}

	userImages, err := h.server.GetUserImagesNotInAlbum(h.ctx, claims.ID, albumID)
	if err != nil {
		http.Error(w, "Failed to load available images", http.StatusInternalServerError)
		return
	}

	tmpl, _ := template.ParseFiles("web/templates/layout.html", "web/templates/album.html")

	data := struct {
		Album      model.Album
		Images     []model.Image
		UserImages []model.Image
	}{
		Album:      model.Album{ID: albumID}, // если хочешь — можно подгрузить полностью
		Images:     albumImages,
		UserImages: userImages,
	}

	fmt.Printf("\n user images %v\n", userImages)

	fmt.Printf("\n user id %v --- \nalbum id: %v\n", claims, albumID)

	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)
	layoutData.Content = data
	layoutData.Title = "Albums Images"

	tmpl.ExecuteTemplate(w, "layout", layoutData)
}

func (h *handler) handleAddToAlbum(w http.ResponseWriter, r *http.Request) {
	albumIDStr := chi.URLParam(r, "id")
	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid album ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Читаем ID изображений
	imageIDsStr := r.Form["image_ids"]
	if len(imageIDsStr) == 0 {
		http.Redirect(w, r, "/albums/"+albumIDStr, http.StatusSeeOther)
		return
	}

	var imageIDs []int64
	for _, idStr := range imageIDsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			imageIDs = append(imageIDs, id)
		}
	}

	if len(imageIDs) == 0 {
		http.Redirect(w, r, "/albums/"+albumIDStr, http.StatusSeeOther)
		return
	}

	// Добавляем изображения в альбом
	err = h.server.AddToAlbum(h.ctx, albumID, imageIDs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add images to album: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/albums/"+albumIDStr, http.StatusSeeOther)
}

func (h *handler) handleRemoveFromAlbum(w http.ResponseWriter, r *http.Request) {
	albumIDStr := chi.URLParam(r, "id")
	imageIDStr := chi.URLParam(r, "image_id")

	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid album ID", http.StatusBadRequest)
		return
	}

	imageID, err := strconv.ParseInt(imageIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	err = h.server.DeleteImageFromAlbum(h.ctx, albumID, imageID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove image from album: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/albums/"+albumIDStr, http.StatusSeeOther)
}

func (h *handler) handleDeleteAlbum(w http.ResponseWriter, r *http.Request) {
	albumIDStr := chi.URLParam(r, "id")

	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid album ID", http.StatusBadRequest)
		return
	}

	err = h.server.DeleteAlbum(h.ctx, albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete album: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/albums", http.StatusSeeOther)
}
