package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type OEmbedResponse struct {
	ThumbnailURL string `json:"thumbnail_url"`
	AuthorName   string `json:"author_name"`
}

func ProcessIGLink(igUrl string) (string, error) {
	// หมายเหตุ: คุณต้องมี Access Token จาก Facebook Developer App
	accessToken := "YOUR_FACEBOOK_APP_ACCESS_TOKEN"
	oEmbedUrl := fmt.Sprintf("https://graph.facebook.com/v18.0/instagram_oembed?url=%s&access_token=%s", igUrl, accessToken)

	// 1. ดึง thumbnail_url จาก oEmbed
	resp, err := http.Get(oEmbedUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("instagram api error: %d", resp.StatusCode)
	}

	var data OEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	// 2. ดาวน์โหลดรูปภาพจาก thumbnail_url
	imgResp, err := http.Get(data.ThumbnailURL)
	if err != nil {
		return "", err
	}
	defer imgResp.Body.Close()

	// 3. บันทึกลงในโฟลเดอร์ /uploads/ig_cache/
	// สร้างชื่อไฟล์จาก ID หรือ URL (ตัวอย่างนี้ใช้ส่วนท้ายของ URL)
	urlParts := strings.Split(strings.TrimSuffix(igUrl, "/"), "/")
	fileName := urlParts[len(urlParts)-1] + ".jpg"

	// ตรวจสอบและสร้างโฟลเดอร์
	cacheDir := "./uploads/ig_cache"
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.MkdirAll(cacheDir, 0755)
	}

	filePath := filepath.Join(cacheDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, imgResp.Body)
	if err != nil {
		return "", err
	}

	// 4. ส่ง Path ภายในเครื่องกลับไปแทน
	return filePath, nil
}
