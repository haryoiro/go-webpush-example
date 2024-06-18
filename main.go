package main

import (
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/labstack/echo/v4"
)

//go:embed key/*.pem
var keyFS embed.FS

var subscriptionsMap map[string]webpush.Subscription

func main() {
	e := echo.New()
	subscriptionsMap = make(map[string]webpush.Subscription)
	e.Static("/static", "assets")
	e.File("/", "public/index.html")
	e.File("/wp.js", "assets/wp.js")

	// ルーターの設定
	e.GET("/webpush/key", getSubscribeKey)
	e.POST("/webpush/subscribe", subscribe)
	e.POST("/webpush/unsubscribe", unsubscribe)
	e.POST("/webpush/notify", sendNotifications)

	e.Logger.Fatal(e.Start(":3000"))
}

func subscribe(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	var subscription webpush.Subscription
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		return err
	}

	_, exists := subscriptionsMap[subscription.Keys.Auth]
	if exists {
		// サブスクリプションが存在する
		return c.JSON(http.StatusOK, map[string]string{"message": "subscription already exists"})
	}

	subscriptionsMap[subscription.Endpoint] = subscription
	return c.JSON(http.StatusOK, map[string]string{"message": "subscription success"})
}

func unsubscribe(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	var subscription webpush.Subscription
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		return err
	}

	_, exists := subscriptionsMap[subscription.Endpoint]
	if !exists {
		// サブスクリプションが存在しない
		return c.JSON(http.StatusOK, map[string]string{"message": "subscription not exists"})
	}

	delete(subscriptionsMap, subscription.Endpoint)
	return c.JSON(http.StatusOK, map[string]string{"message": "unsubscription success"})
}

func getSubscribeKey(c echo.Context) error {
	// 公開鍵を読み込む
	vapidPublicKey, err := keyFS.ReadFile("key/pub.pem")
	if err != nil {
		return err
	}

	return c.Blob(http.StatusOK, "text/plain", vapidPublicKey)
}

type NotificationRequest struct {
	Title   string `form:"title"`
	Message string `form:"message"`
}

func sendNotifications(c echo.Context) error {
	var request NotificationRequest
	if err := c.Bind(&request); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	for _, subscription := range subscriptionsMap {
		if err := sendNotification(&subscription, request.Title, request.Message); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	return c.Redirect(http.StatusFound, "/")
}

type WebPushPayload struct {
	Icon    string `json:"icon"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func NewWebPushPayload(title string, message string) WebPushPayload {
	return WebPushPayload{
		Icon:    "/static/icon.png",
		Title:   title,
		Message: message,
	}
}

func sendNotification(subscription *webpush.Subscription, title string, message string) error {
	webPushPayload := NewWebPushPayload(title, message)
	msg, err := json.Marshal(webPushPayload)
	if err != nil {
		return err
	}

	privateKey, err := keyFS.ReadFile("key/priv.pem")
	if err != nil {
		return err
	}
	publicKey, err := keyFS.ReadFile("key/pub.pem")
	if err != nil {
		return err
	}

	resp, err := webpush.SendNotification([]byte(msg), subscription, &webpush.Options{
		Subscriber:      "info@haryoiro.com",
		VAPIDPublicKey:  decodeKey(publicKey),
		VAPIDPrivateKey: decodeKey(privateKey),
		TTL:             86400,
		Urgency:         webpush.UrgencyNormal,
	})
	if err != nil {
		return err
	}

	resp.Header.Set("Content-Type", "application/octet-stream")
	resp.Header.Set("TTL", "86400")
	resp.Header.Set("Urgency", "normal")
	return nil

}

func decodeKey(publicKeyBytes []byte) string {
	publicKey := string(publicKeyBytes)
	return strings.TrimSpace(publicKey)
}
