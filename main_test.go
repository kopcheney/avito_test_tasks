package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Statistics struct {
	Likes     int `json:"likes"`
	ViewCount int `json:"viewCount"`
	Contacts  int `json:"contacts"`
}

type ARequest struct {
	SellerID   int        `json:"sellerID"`
	Name       string     `json:"name"`
	Price      int        `json:"price"`
	Statistics Statistics `json:"statistics"`
}

type AResponse struct {
	ID         string     `json:"id"`
	SellerId   int        `json:"sellerId"`
	Name       string     `json:"name"`
	Price      int        `json:"price"`
	Statistics Statistics `json:"statistics"`
	CreatedAt  string     `json:"createdAt"`
}

func TestCreate(t *testing.T) {
	sellerID := 123321
	reqBody := ARequest{
		SellerID:   sellerID,
		Name:       "Тестовый Товар",
		Price:      1000,
		Statistics: Statistics{Likes: 1, ViewCount: 2, Contacts: 3},
	}

	jsonData, err := json.Marshal(reqBody)
	assert.NoError(t, err, "ошибка преобразования JSON")

	url := "https://qa-internship.avito.com"
	httpReq, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(jsonData),
	)
	assert.NoError(t, err, "Ошибка при создании запроса")

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
}

func TestAByID(t *testing.T) {
	sellerID := 123321

	reqBody := ARequest{
		SellerID:   sellerID,
		Name:       "Тестовое объявление для GET",
		Price:      2500,
		Statistics: Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
	}

	jsonData, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	postURL := "https://qa-internship.avito.com/api/1/item"
	postReq, err := http.NewRequest(http.MethodPost, postURL, bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Accept", "application/json")

	client := &http.Client{}

	resp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	t.Logf("POST /item → Статус: %d, Тело: %s", resp.StatusCode, string(bodyBytes))

	var statusResp map[string]string
	err = json.Unmarshal(bodyBytes, &statusResp)
	assert.NoError(t, err)

	statusMsg, exists := statusResp["status"]
	assert.True(t, exists, "Поле 'status' отсутствует в ответе")

	parts := strings.Split(statusMsg, " - ")
	assert.Len(t, parts, 2, "Неверный формат сообщения 'status'")
	itemID := parts[1]
	assert.NotEmpty(t, itemID, "ID не может быть пустым")
	t.Logf("Извлечённый itemID: %s", itemID)

	getURL := "https://qa-internship.avito.com/api/1/item/" + itemID
	getReq, err := http.NewRequest(http.MethodGet, getURL, nil)
	assert.NoError(t, err)
	getReq.Header.Set("Accept", "application/json")

	getResp, err := client.Do(getReq)
	assert.NoError(t, err)
	defer getResp.Body.Close()

	getBodyBytes, err := io.ReadAll(getResp.Body)
	assert.NoError(t, err)
	t.Logf("GET /item/%s → Статус: %d, Тело: %s", itemID, getResp.StatusCode, string(getBodyBytes))

	var announcements []AResponse
	err = json.Unmarshal(getBodyBytes, &announcements)
	assert.NoError(t, err)

	assert.NotEmpty(t, announcements, "Ответ должен содержать хотя бы одно объявление")

	assert.Equal(t, itemID, announcements[0].ID, "ID в ответе не совпадает с запрошенным")
}

func TestGetABySellerID(t *testing.T) {
	sellerID := 123321

	reqBody := ARequest{
		SellerID:   sellerID,
		Name:       "Тестовое объявление для sellerID",
		Price:      3000,
		Statistics: Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
	}

	jsonData, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	postURL := "https://qa-internship.avito.com/api/1/item"
	postReq, err := http.NewRequest(http.MethodPost, postURL, bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var createdAnnouncement AResponse
	err = json.NewDecoder(resp.Body).Decode(&createdAnnouncement)
	assert.NoError(t, err)

	getURL := fmt.Sprintf("https://qa-internship.avito.com/api/1/%d/item", sellerID)
	getReq, err := http.NewRequest(http.MethodGet, getURL, nil)
	assert.NoError(t, err)
	getReq.Header.Set("Accept", "application/json")

	getResp, err := client.Do(getReq)
	assert.NoError(t, err)
	defer getResp.Body.Close()

	var announcements []AResponse
	err = json.NewDecoder(getResp.Body).Decode(&announcements)
	assert.NoError(t, err)

	assert.NotEmpty(t, announcements, "Список объявлений не должен быть пустым")

	found := false
	for _, ann := range announcements {
		if ann.ID == createdAnnouncement.ID {
			found = true
			assert.Equal(t, createdAnnouncement.Name, ann.Name)
			assert.Equal(t, createdAnnouncement.Price, ann.Price)
			break
		}
	}
	assert.True(t, found, "Созданное объявление не найдено в списке по sellerID")
}

func TestSByItemID(t *testing.T) {
	sellerID := 123321

	initialStats := Statistics{Likes: 1, ViewCount: 1, Contacts: 1}
	reqBody := ARequest{
		SellerID:   sellerID,
		Name:       "Объявление для статистики",
		Price:      4000,
		Statistics: initialStats,
	}

	jsonData, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	postURL := "https://qa-internship.avito.com/api/1/item"
	postReq, err := http.NewRequest(http.MethodPost, postURL, bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(postReq)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var createdAnnouncement AResponse
	err = json.NewDecoder(resp.Body).Decode(&createdAnnouncement)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdAnnouncement.ID)

	itemID := createdAnnouncement.ID
	t.Logf("Создан itemID: %s", itemID)

	statURL := fmt.Sprintf("https://qa-internship.avito.com/api/1/item/%s/statistic", itemID)
	statReq, err := http.NewRequest(http.MethodGet, statURL, nil)
	assert.NoError(t, err)
	statReq.Header.Set("Accept", "application/json")

	statResp, err := client.Do(statReq)
	assert.NoError(t, err)
	defer statResp.Body.Close()

	bodyBytes, err := io.ReadAll(statResp.Body)
	assert.NoError(t, err)
	t.Logf("Тело статистики: %s", string(bodyBytes))

	var statsList []Statistics
	err = json.Unmarshal(bodyBytes, &statsList)

	if err != nil {
		var singleStat Statistics
		err2 := json.Unmarshal(bodyBytes, &singleStat)
		if err2 != nil {
			t.Fatalf("Не удалось распарсить статистику ни как массив, ни как объект: %v", err2)
		}
		statsList = []Statistics{singleStat}
	}

	assert.NotEmpty(t, statsList, "Статистика не должна быть пустой")

	stat := statsList[0]
	assert.GreaterOrEqual(t, stat.Likes, 0, "likes должно быть >= 0")
	assert.GreaterOrEqual(t, stat.ViewCount, 0, "viewCount должно быть >= 0")
	assert.GreaterOrEqual(t, stat.Contacts, 0, "contacts должно быть >= 0")
}
