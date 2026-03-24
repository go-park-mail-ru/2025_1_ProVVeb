package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
	"github.com/sirupsen/logrus"
)

type SessionHandler struct {
	LoginUC        usecase.UserLogIn
	CheckSessionUC usecase.UserCheckSession
	LogoutUC       usecase.UserLogOut
	Logger         *logger.LogrusLogger
}

type SubHandler struct {
	AddSubUC       usecase.AddSubscription
	UpdateBorderUC usecase.UpdateBorder

	Logger *logger.LogrusLogger
}

type UserHandler struct {
	SignupUC     usecase.UserSignUp
	DeleteUserUC usecase.DeleteUser
	GetParamsUC  usecase.UserGetParams
	GetPremiumUC usecase.GetPremium
	Logger       *logger.LogrusLogger
}

type ComplaintHandler struct {
	GetComplaintsUC    usecase.GetComplaint
	CreateComplateUC   usecase.CreateComplaint
	FindCompaintUC     usecase.FindComplaint
	DeleteComplaintsUC usecase.DeleteComplaint
	HandleComplaintUC  usecase.HandleComplaint
	GetStatisticsUC    usecase.GetStatisticsCompl

	GetAdminUC usecase.GetAdmin

	Logger *logger.LogrusLogger
}

type ProfilesHandler struct {
	DeleteImageUC         usecase.DeleteStatic
	GetProfileImagesUC    usecase.GetUserPhoto
	GetProfileMatchesUC   usecase.GetProfileMatches
	GetProfileUC          usecase.GetProfile
	GetProfilesUC         usecase.GetProfilesForUser
	SetProfilesLikeUC     usecase.ProfileSetLike
	UpdateProfileUC       usecase.ProfileUpdate
	UpdateProfileImagesUC usecase.StaticUpload
	SearchProfileUC       usecase.SearchProfiles
	AddNotificationUC     usecase.AddNotification
	Subscriber            *redis.Client
	GetAdminUC            usecase.GetAdmin
	GetRecommendationsUC  usecase.GetRecommendations
	GetProfileStatsUC     usecase.GetProfileStats

	Logger *logger.LogrusLogger
}

type QueryHandler struct {
	GetActiveQueriesUC   usecase.GetActiveQueries
	StoreUserAnswerUC    usecase.StoreUserAnswer
	GetAnswersForUserUC  usecase.GetAnswersForUser
	GetAnswersForQueryUC usecase.GetAnswersForQuery
	FindQueryUC          usecase.FindQuery
	DeleteQueryUC        usecase.DeleteAnswer
	GetStatisticsUC      usecase.GetStatistics
	GetAdminUC           usecase.GetAdmin
	Logger               *logger.LogrusLogger
}

type MessageHandler struct {
	GetParticipantsUC usecase.GetChatParticipants
	GetChatsUC        usecase.GetChats
	CreateChatUC      usecase.CreateChat
	DeleteChatUC      usecase.DeleteChat

	GetMessagesUC          usecase.GetMessages
	DeleteMessageUC        usecase.DeleteMessage
	CreateMessagesUC       usecase.CreateMessages
	GetMessagesFromCacheUC usecase.GetMessagesFromCache
	UpdateMessageStatusUC  usecase.UpdateMessageStatus

	AddNotificationUC usecase.AddNotification

	Subscriber *redis.Client

	Logger *logger.LogrusLogger
}

type NotificationsHandler struct {
	GetNotificationsUC         usecase.GetNotifications
	UpdateNotificationStatusUC usecase.UpdateNotificationStatus
	DeleteNotificationUC       usecase.DeleteNotification

	GetCurrentNotificationsUC usecase.GetCurrentNotifications
	Subscriber                *redis.Client
	Logger                    *logger.LogrusLogger
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (mh *NotificationsHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	mh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("request started")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		mh.Logger.WithFields(&logrus.Fields{
			"error": "failed to get userID from context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Failed to establish WebSocket connection"},
		)
	}
	defer conn.Close()

	notifications, err := mh.GetNotificationsUC.GetNotifications(int(profileId))
	if err != nil {
		mh.Logger.Error("Failed to load initial notifications: ", err)
		conn.WriteJSON(map[string]interface{}{"error": fmt.Sprintf("Failed to load initial notifications %v", err)})
		return
	}
	conn.WriteJSON(map[string]interface{}{"type": "init_notifications", "notifications": notifications})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	channelName := fmt.Sprintf("user:%d notifications", profileId)
	pubsub := mh.Subscriber.Subscribe(ctx, channelName)
	defer pubsub.Close()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			case <-pubsub.Channel():
				newNotifications, err := mh.GetCurrentNotificationsUC.GetCurrentNotifications(int(profileId))
				if err != nil {
					mh.Logger.Error("Failed to get messages from cache (ticker):", err)
					conn.WriteJSON(map[string]interface{}{"error": "Failed to get messages"})
					continue
				}
				if len(newNotifications) > 0 {
					conn.WriteJSON(map[string]interface{}{"type": "new_notifications", "notifications": newNotifications})
				}
			}
		}
	}()

	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			mh.Logger.Error("Error reading message from WebSocket: ", err)
			conn.WriteJSON(map[string]interface{}{"error": "Error reading message"})
			close(done)
			break
		}

		var wsMessage model.WSMessage
		if err := easyjson.Unmarshal(msgData, &wsMessage); err != nil {
			mh.Logger.Error("Failed to unmarshal WSMessage: ", err)
			conn.WriteJSON(map[string]interface{}{"error": "Invalid message format"})
			continue
		}

		switch wsMessage.Type {
		case "sendFlowers":
			messageReceived.WithLabelValues().Inc()
			var payload model.FlowersPayload
			if err := easyjson.Unmarshal(wsMessage.Payload, &payload); err != nil {
				mh.Logger.Error("Failed to unmarshal payload: ", err)
				conn.WriteJSON(map[string]interface{}{"error": "Invalid payload"})
				break
			}
			go func(payload model.FlowersPayload) {
				notif := model.NotificationSend{
					NotifType: "flowers",
					Content:   fmt.Sprintf("User %d sent you flowers!", profileId),
					Read:      0,
				}
				data, _ := json.Marshal(notif)
				channel := fmt.Sprintf("user:%d notifications", payload.UserID)
				err := mh.Subscriber.Publish(context.Background(), channel, data).Err()
				if err != nil {
					mh.Logger.Error("Failed to publish flowers notification: ", err)
					conn.WriteJSON(map[string]interface{}{"error": "Failed to notify"})
					return
				}
				conn.WriteJSON(map[string]interface{}{"type": "SentFlowersTo", "user": payload.UserID})

				redisKey := fmt.Sprintf("CACHE:user:%dnotifications", payload.UserID)

				jsonNotif, err := json.Marshal(notif)
				if err != nil {
					return
				}

				pipe := mh.Subscriber.TxPipeline()
				pipe.LPush(context.Background(), redisKey, jsonNotif)
				pipe.Expire(context.Background(), redisKey, 30*time.Minute)
				_, _ = pipe.Exec(context.Background())

			}(payload)

		case "delete":
			messageReceived.WithLabelValues().Inc()
			var payload model.DeleteNotifPayload
			if err := easyjson.Unmarshal(wsMessage.Payload, &payload); err != nil {
				mh.Logger.Error("Failed to unmarshal ReadPayload: ", err)
				conn.WriteJSON(map[string]interface{}{"error": "Invalid read payload"})
				break
			}
			go func(payload model.DeleteNotifPayload) {
				err := mh.DeleteNotificationUC.DeleteNotifications(payload.NotifID, int(profileId))
				if err != nil {
					mh.Logger.Error("Failed to update message status: ", err)
					conn.WriteJSON(map[string]interface{}{"error": "Failed to update message status"})
					return
				}
				conn.WriteJSON(map[string]interface{}{"type": "status_updated", "user": profileId})
			}(payload)

		case "read":
			var payload model.ReadNotifPayload
			if err := easyjson.Unmarshal(wsMessage.Payload, &payload); err != nil {
				mh.Logger.Error("Failed to unmarshal payload: ", err)
				conn.WriteJSON(map[string]interface{}{"error": "Invalid payload"})
				break
			}
			messageReceived.WithLabelValues().Inc()
			go func(payload model.ReadNotifPayload) {
				err := mh.UpdateNotificationStatusUC.UpdateNotificatons(int(profileId), payload.NotifType)
				if err != nil {
					mh.Logger.Error("Failed to update message status: ", err)
					conn.WriteJSON(map[string]interface{}{"error": "Failed to update message status"})
					return
				}
				conn.WriteJSON(map[string]interface{}{"type": "status_updated", "user": profileId})
			}(payload)

		default:
			mh.Logger.Warn("Unknown action: ", wsMessage.Type)
			conn.WriteJSON(map[string]interface{}{"error": "Unknown action type"})
		}
	}
}

func (mh *MessageHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatIDStr, ok := vars["chat_id"]
	if !ok {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Missing chat_id in URL"},
		)
		return
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid chat_id format"},
		)
		return
	}

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		mh.Logger.WithFields(&logrus.Fields{
			"error": "failed to get userID from context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Failed to establish WebSocket connection"},
		)
		return
	}
	defer conn.Close()

	first, second, err := mh.GetParticipantsUC.GetChatParticipants(chatID)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Failed to get chat participants"},
		)
		return
	}

	if profileId != uint32(first) && profileId != uint32(second) {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	messages, err := mh.GetMessagesUC.GetMessages(chatID)
	if err != nil {
		mh.Logger.Error("Failed to load initial messages: ", err)
		conn.WriteJSON(map[string]interface{}{"error": "Failed to load initial messages"})
		return
	}
	new_messages_first, err := mh.GetMessagesFromCacheUC.GetMessages(chatID, first)
	if err != nil {
		mh.Logger.Error("Failed to load initial messages: ", err)
		conn.WriteJSON(map[string]interface{}{"error": "Failed to load initial messages"})
		return
	}
	new_messages_second, err := mh.GetMessagesFromCacheUC.GetMessages(chatID, second)
	if err != nil {
		mh.Logger.Error("Failed to load initial messages: ", err)
		conn.WriteJSON(map[string]interface{}{"error": "Failed to load initial messages"})
		return
	}

	msgMap := make(map[int]model.Message)

	for _, m := range messages {
		msgMap[m.MessageID] = m
	}
	for _, m := range append(new_messages_first, new_messages_second...) {
		if _, exists := msgMap[m.MessageID]; !exists {
			msgMap[m.MessageID] = m
		}
	}
	var allMessages []model.Message
	for _, m := range msgMap {
		allMessages = append(allMessages, m)
	}

	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].CreatedAt.Before(allMessages[j].CreatedAt)
	})

	conn.WriteJSON(map[string]interface{}{"type": "init_messages", "messages": allMessages})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	channelName := fmt.Sprintf("user:%d chat:%d messages", profileId, chatID)
	pubsub := mh.Subscriber.Subscribe(ctx, channelName)
	defer pubsub.Close()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			case <-pubsub.Channel():
				newMessages, err := mh.GetMessagesFromCacheUC.GetMessages(chatID, int(profileId))
				if err != nil {
					mh.Logger.Error("Failed to get messages from cache (ticker): ", err)
					conn.WriteJSON(map[string]interface{}{"error": "Failed to get messages"})
					continue
				}
				if len(newMessages) > 0 {
					conn.WriteJSON(map[string]interface{}{"type": "new_messages", "messages": newMessages})
				}
			}
		}
	}()

	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			mh.Logger.Error("Error reading message from WebSocket: ", err)
			conn.WriteJSON(map[string]interface{}{"error": "Error reading message"})
			close(done)
			break
		}

		var wsMessage model.WSMessage
		if err := easyjson.Unmarshal(msgData, &wsMessage); err != nil {
			mh.Logger.Error("Failed to unmarshal WSMessage: ", err)
			conn.WriteJSON(map[string]interface{}{"error": "Invalid message format"})
			continue
		}

		switch wsMessage.Type {
		case "create":
			messageSent.WithLabelValues().Inc()
			var payload model.CreatePayload
			if err := easyjson.Unmarshal(wsMessage.Payload, &payload); err != nil {
				mh.Logger.Error("Failed to unmarshal CreatePayload: ", err)
				conn.WriteJSON(map[string]interface{}{"error": "Invalid create payload"})
				break
			}
			if ((payload.UserID != first) && (payload.UserID != second)) || (payload.ChatID != chatID) {
				MakeEasyJSONResponse(w, http.StatusUnauthorized,
					&model.ErrorResponse{Message: "You don't have access"},
				)
				break
			}
			recieverID := first
			if payload.UserID == first {
				recieverID = second
			}

			notif := model.NotificationSend{
				NotifType: "message",
				Content:   fmt.Sprintf("User %d sent you a message!", payload.UserID),
				Read:      0,
			}

			err := mh.AddNotificationUC.AddNotification(recieverID, notif)
			if err != nil {
				fmt.Println(err)
				mh.Logger.Error("Failed to save notification: ", err)
				conn.WriteJSON(map[string]interface{}{"error": "Failed to notify"})
				return
			}

			go func(payload model.CreatePayload) {
				messageID, err := mh.CreateMessagesUC.CreateMessages(payload.ChatID, payload.UserID, payload.Content)
				if err != nil {
					mh.Logger.Error("Failed to create message: ", err)
					conn.WriteJSON(map[string]interface{}{"error": "Failed to create message"})
					return
				}
				conn.WriteJSON(map[string]interface{}{"type": "created", "message_id": messageID})
			}(payload)

		case "delete":
			messageSent.WithLabelValues().Inc()
			var payload model.DeletePayload
			if err := easyjson.Unmarshal(wsMessage.Payload, &payload); err != nil {
				mh.Logger.Error("Failed to unmarshal DeletePayload: ", err)
				conn.WriteJSON(map[string]interface{}{"error": "Invalid delete payload"})
				break
			}
			if payload.ChatID != chatID {
				MakeEasyJSONResponse(w, http.StatusUnauthorized,
					&model.ErrorResponse{Message: "You don't have access"},
				)
				break
			}
			go func(payload model.DeletePayload) {
				err := mh.DeleteMessageUC.DeleteMessage(payload.MessageID, payload.ChatID)
				if err != nil {
					mh.Logger.Error("Failed to delete message: ", err)
					conn.WriteJSON(map[string]interface{}{"error": "Failed to delete message"})
					return
				}
				conn.WriteJSON(map[string]interface{}{"type": "deleted", "message_id": payload.MessageID})
			}(payload)

		case "get":
			messageReceived.WithLabelValues().Inc()
			newMessages, err := mh.GetMessagesFromCacheUC.GetMessages(chatID, int(profileId))
			if err != nil {
				mh.Logger.Error("Failed to get messages from cache: ", err)
				conn.WriteJSON(map[string]interface{}{"error": "Failed to get messages"})
				break
			}
			conn.WriteJSON(map[string]interface{}{"type": "new_messages", "messages": newMessages})

		case "read":
			messageReceived.WithLabelValues().Inc()
			var payload model.ReadPayload
			if err := easyjson.Unmarshal(wsMessage.Payload, &payload); err != nil {
				mh.Logger.Error("Failed to unmarshal ReadPayload: ", err)
				conn.WriteJSON(map[string]interface{}{"error": "Invalid read payload"})
				break
			}
			if payload.ChatID != chatID {
				MakeEasyJSONResponse(w, http.StatusUnauthorized,
					&model.ErrorResponse{Message: "You don't have access"},
				)
				break
			}
			go func(payload model.ReadPayload) {
				err := mh.UpdateMessageStatusUC.UpdateMessageStatus(payload.ChatID, int(profileId))
				if err != nil {
					mh.Logger.Error("Failed to update message status: ", err)
					conn.WriteJSON(map[string]interface{}{"error": "Failed to update message status"})
					return
				}
				conn.WriteJSON(map[string]interface{}{"type": "status_updated", "chat": payload.ChatID})
			}(payload)

		default:
			mh.Logger.Warn("Unknown action: ", wsMessage.Type)
			conn.WriteJSON(map[string]interface{}{"error": "Unknown action type"})
		}
	}
}

func (mh *MessageHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	mh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("start processing CreateChat request")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		mh.Logger.WithFields(&logrus.Fields{
			"error": "failed to get userID from context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid request body"},
		)
		return
	}

	var req model.CreateChatRequest
	if err := req.UnmarshalJSON(body); err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		return
	}

	if (req.FristID != int(profileId)) && (req.SecondID != int(profileId)) {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "Unauthorized"},
		)
		return
	}

	chatID, err := mh.CreateChatUC.CreateChat(req.FristID, req.SecondID)
	if err != nil {
		mh.Logger.WithFields(&logrus.Fields{
			"FirstID":  req.FristID,
			"SecondID": req.SecondID,
			"error":    err.Error(),
		}).Error("failed to create chat")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error creating chat: %v", err)},
		)
		return
	}

	mh.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
		"chatID":     chatID,
	}).Info("successfully created chat")

	MakeEasyJSONResponse(w, http.StatusCreated,
		&model.ErrorResponse{Message: "Chat created"},
	)

}

func (mh *MessageHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	mh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("start processing GetChats request")
	messageChatsViews.WithLabelValues().Inc()

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		mh.Logger.WithFields(&logrus.Fields{
			"error": "failed to get userID from context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	chats, err := mh.GetChatsUC.GetChats(int(profileId))
	if err != nil {
		mh.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get chats")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting chats: %v", err)},
		)

		return
	}

	mh.Logger.WithFields(&logrus.Fields{
		"profile_id":  profileId,
		"chats_count": len(chats),
	}).Info("successfully retrieved chats")

	MakeEasyJSONResponse(w, http.StatusOK, model.ChatsResponse{Chats: chats})
}

func (mh *MessageHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	mh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("start processing DeleteChat request")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		mh.Logger.WithFields(&logrus.Fields{
			"error": "failed to get userID from context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid request body"},
		)
		return
	}

	var req model.DeleteChatRequest
	if err := req.UnmarshalJSON(body); err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		return
	}

	if (req.FristID != int(profileId)) && (req.SecondID != int(profileId)) {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "Unauthorized"},
		)
		return
	}

	err = mh.DeleteChatUC.DeleteChat(req.FristID, req.SecondID)
	if err != nil {
		mh.Logger.WithFields(&logrus.Fields{
			"FirstID":  req.FristID,
			"SecondID": req.SecondID,
			"error":    err.Error(),
		}).Error("failed to delete chat")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error creating chat: %v", err)},
		)
		return
	}

	mh.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Info("successfully deleted chat")

	MakeEasyJSONResponse(w, http.StatusCreated,
		&model.ErrorResponse{Message: "Chat deleted"},
	)
}

func (ph *ProfilesHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("start processing UpdateProfile request")

	profileUpdated.WithLabelValues("update profile").Inc()
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "failed to get userID from context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"error":      err.Error(),
			"profile_id": profileId,
		}).Warn("failed to read request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid request body"},
		)
		return
	}

	var profile model.Profile
	if err := profile.UnmarshalJSON(body); err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"error":      err.Error(),
			"profile_id": profileId,
		}).Warn("failed to unmarshal profile")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		return
	}

	profile.ProfileId = int(profileId)

	table_profile, err := ph.GetProfileUC.GetProfile(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get profile from database")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting profile: %v", err)},
		)
		return
	}

	err = ph.UpdateProfileUC.UpdateProfile(profile, table_profile, int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id":   profileId,
			"error":        err.Error(),
			"profile_data": profile,
		}).Error("failed to update profile")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error updating profile: %v", err)},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Info("profile updated successfully")

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: "Updated"},
	)
}

func (ph *ProfilesHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("GetMatches request started")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Debug("attempting to get matches")

	profiles, err := ph.GetProfileMatchesUC.GetMatches(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get matches")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting profile: %v", err)},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id":    profileId,
		"matches_count": len(profiles),
	}).Info("successfully retrieved matches")

	MakeEasyJSONResponse(w, http.StatusOK, model.ProfileResponse{Profiles: profiles})
}

func (ph *ProfilesHandler) SearchProfiles(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("SearchProfiles request started")

	searchPerformed.WithLabelValues("search profiles").Inc()
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized profiles access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"requester_id": profileId,
	}).Debug("attempting to get profiles list")

	var input model.SearchProfileRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Warn("failed to read request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid request body"},
		)
		return
	}

	if err := input.UnmarshalJSON(body); err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Warn("failed to unmarshal search profile request")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		return
	}
	profiles, err := ph.SearchProfileUC.GetSearchProfiles(int(profileId), input)
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"requester_id": profileId,
			"error":        err.Error(),
		}).Error("failed to get profiles list")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting profile: %v", err)},
		)
		return
	}

	if len(profiles) == 0 {
		MakeEasyJSONResponse(w, http.StatusAccepted,
			&model.ErrorResponse{Message: "There are no profiles"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"requester_id":   profileId,
		"profiles_count": len(profiles),
	}).Info("profiles list retrieved successfully")

	MakeEasyJSONResponse(w, http.StatusOK, model.FoundProfileResponse{Profiles: profiles})
}

func (ph *ProfilesHandler) SetLike(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("SetLike request started")

	likeSet.WithLabelValues("set like").Inc()
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	IsPremiumRaw := r.Context().Value(isPremiumKey)
	IsPremium, _ := IsPremiumRaw.(bool)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Warn("failed to read like request body")
		MakeEasyJSONResponse(w, http.StatusBadRequest, &model.ErrorResponse{Message: "Invalid request body"})
		return
	}

	var input model.SetLike
	if err := input.UnmarshalJSON(body); err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Warn("failed to unmarshal like request body")
		MakeEasyJSONResponse(w, http.StatusBadRequest, &model.ErrorResponse{Message: "Invalid JSON"})
		return
	}

	likeFrom := input.LikeFrom
	likeTo := input.LikeTo
	status := input.Status

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
		"like_from":  likeFrom,
		"like_to":    likeTo,
		"status":     status,
	}).Debug("processing like action")

	if likeTo == likeFrom {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
		}).Warn("attempt to like oneself")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Please don't like yourself"},
		)
		return
	}

	if (!IsPremium) && (status == 3) {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"like_from":  likeFrom,
		}).Warn("no premium")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "You cannot use superlike with no subscription"},
		)
		return
	}

	if int(profileId) != likeFrom {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"like_from":  likeFrom,
		}).Warn("unauthorized like attempt")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "You are unauthorized to like this user"},
		)
		return
	}

	like_id, err := ph.SetProfilesLikeUC.SetLike(likeFrom, likeTo, status)
	if (like_id == 0) && (err == nil) {
		ph.Logger.WithFields(&logrus.Fields{
			"like_from": likeFrom,
			"like_to":   likeTo,
		}).Info("duplicate like detected")

		MakeEasyJSONResponse(w, http.StatusConflict,
			&model.ErrorResponse{Message: "Already liked"},
		)
		return
	}
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"like_from": likeFrom,
			"like_to":   likeTo,
			"error":     err.Error(),
		}).Error("failed to set like")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting like: %v", err)},
		)
		return
	}

	if like_id == -1 {
		recieverID := likeTo

		notif := model.NotificationSend{
			NotifType: "match",
			Content:   fmt.Sprintf("You have matched with user %d!", likeFrom),
			Read:      0,
		}

		err := ph.AddNotificationUC.AddNotification(recieverID, notif)
		if err != nil {
			ph.Logger.Error("Failed to save notification: ", err)
			return
		}

		recieverID = likeFrom

		notif = model.NotificationSend{
			NotifType: "match",
			Content:   fmt.Sprintf("You have natched with user %d!", likeTo),
			Read:      0,
		}

		err = ph.AddNotificationUC.AddNotification(recieverID, notif)
		if err != nil {
			ph.Logger.Error("Failed to save notification: ", err)
			return
		}
	}

	redisKey := fmt.Sprintf("cached_profiles:%d", profileId)

	cachedData, err := ph.Subscriber.Get(context.Background(), redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return
		}
		ph.Logger.WithError(err).Error("failed to get cached profiles")
		return
	}

	var profiles []model.Profile
	if err := json.Unmarshal([]byte(cachedData), &profiles); err != nil {
		ph.Logger.WithError(err).Error("failed to unmarshal cached profiles")
		return
	}

	filtered := make([]model.Profile, 0, len(profiles))
	for _, p := range profiles {
		if p.ProfileId != likeTo {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		err = ph.Subscriber.Del(context.Background(), redisKey).Err()
		if err != nil {
			ph.Logger.WithError(err).Error("failed to update cached profiles in redis")
		}
		return
	}

	newData, err := json.Marshal(filtered)
	if err != nil {
		ph.Logger.WithError(err).Error("failed to marshal updated profiles")
		return
	}

	err = ph.Subscriber.Set(context.Background(), redisKey, newData, 30*time.Minute).Err()
	if err != nil {
		ph.Logger.WithError(err).Error("failed to update cached profiles in redis")
	}

	ph.Logger.WithFields(&logrus.Fields{
		"like_id":   like_id,
		"like_from": likeFrom,
		"like_to":   likeTo,
		"status":    status,
	}).Info("like successfully processed")

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: "Liked"},
	)
}

func CreateCookies(session model.Session) (*model.Cookie, error) {
	cookie := &model.Cookie{
		Name:     "session_id",
		Value:    session.SessionId,
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(12 * time.Hour),
		Path:     "/",
	}
	return cookie, nil
}

func (ph *ProfilesHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":       r.Method,
		"path":         r.URL.Path,
		"request_id":   r.Header.Get("request_id"),
		"content_type": r.Header.Get("Content-Type"),
	}).Info("UploadPhoto request started")

	photoUploaded.WithLabelValues("upload photo").Inc()
	sanitizer := bluemonday.UGCPolicy()
	var maxMemory int64 = model.MaxFileSize
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized upload attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"user_id":    user_id,
		"max_memory": maxMemory,
	}).Debug("parsing multipart form")

	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Warn("failed to parse multipart form")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("Invalid multipart form: %v", err)},
		)
		return
	}

	form := r.MultipartForm
	files := form.File["images"]

	if len(files) == 0 {
		ph.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
		}).Warn("no files in 'images' field")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "No files in 'images' field"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"user_id":       user_id,
		"files_count":   len(files),
		"allowed_types": allowedTypes,
	}).Info("starting files processing")

	var (
		failedUploads  []string
		successUploads []string
	)

	for _, fileHeader := range files {
		fileName := fileHeader.Filename
		fileSize := fileHeader.Size
		contentType := fileHeader.Header.Get("Content-Type")

		ph.Logger.WithFields(&logrus.Fields{
			"file_name":    fileName,
			"file_size":    fileSize,
			"content_type": contentType,
		}).Debug("processing file")

		file, err := fileHeader.Open()
		if err != nil {
			ph.Logger.WithFields(&logrus.Fields{
				"file_name": fileName,
				"error":     err.Error(),
			}).Warn("failed to open file")

			failedUploads = append(failedUploads, fileName)
			continue
		}
		defer file.Close()

		sanitizedType := sanitizer.Sanitize(contentType)
		if !allowedTypes[sanitizedType] {
			ph.Logger.WithFields(&logrus.Fields{
				"file_name":    fileName,
				"content_type": sanitizedType,
			}).Warn("unsupported file type")

			failedUploads = append(failedUploads, fileName+" (unsupported type)")
			continue
		}

		buf, err := io.ReadAll(file)
		if err != nil {
			ph.Logger.WithFields(&logrus.Fields{
				"file_name": fileName,
				"error":     err.Error(),
			}).Warn("failed to read file content")

			failedUploads = append(failedUploads, fileName+" (read error)")
			continue
		}

		filename := fmt.Sprintf("/%d_%d_%s", user_id, time.Now().UnixNano(), fileName)

		ph.Logger.WithFields(&logrus.Fields{
			"user_id":   user_id,
			"file_name": filename,
			"data_size": len(buf),
		}).Debug("uploading file to storage")

		err = ph.UpdateProfileImagesUC.UploadUserPhoto(int(user_id), buf, filename, sanitizedType)
		if err != nil {
			ph.Logger.WithFields(&logrus.Fields{
				"user_id":   user_id,
				"file_name": filename,
				"error":     err.Error(),
			}).Error("failed to upload file")

			failedUploads = append(failedUploads, fileName+" (upload error)")
			continue
		}

		successUploads = append(successUploads, filename)
		ph.Logger.WithFields(&logrus.Fields{
			"user_id":   user_id,
			"file_name": filename,
		}).Info("file uploaded successfully")
	}

	ph.Logger.WithFields(&logrus.Fields{
		"user_id":        user_id,
		"total_files":    len(files),
		"success_count":  len(successUploads),
		"failed_count":   len(failedUploads),
		"failed_uploads": failedUploads,
	}).Info("files processing completed")

	if len(failedUploads) != 0 {
		if len(successUploads) > 0 {
			ph.Logger.WithFields(&logrus.Fields{
				"user_id":       user_id,
				"success_count": len(successUploads),
				"failed_count":  len(failedUploads),
			}).Warn("partial upload failure")
		} else {
			ph.Logger.WithFields(&logrus.Fields{
				"user_id":      user_id,
				"failed_count": len(failedUploads),
			}).Error("all files failed to upload")
		}

		MakeEasyJSONResponse(w, http.StatusInternalServerError, &model.UploadResponse{
			Message:          "Some uploads failed",
			SucessfulUploads: failedUploads,
		})
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"user_id":        user_id,
		"uploaded_files": successUploads,
	}).Info("all files uploaded successfully")

	MakeEasyJSONResponse(w, http.StatusOK, &model.UploadResponse{
		Message:          "All files uploaded",
		SucessfulUploads: successUploads,
	})
}

func (sh *SessionHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	sh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("Login request started")

	sanitizer := bluemonday.UGCPolicy()
	var input model.LoginRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Warn("failed to read login request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Failed to read request body"},
		)
		loginAttempts.WithLabelValues("false").Inc()
		return
	}

	if err := input.UnmarshalJSON(body); err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Warn("failed to decode login request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		loginAttempts.WithLabelValues("false").Inc()
		return
	}

	input.Login = sanitizer.Sanitize(input.Login)
	input.Password = sanitizer.Sanitize(input.Password)

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Cannot parse IP", http.StatusInternalServerError)
		return
	}

	blockTime, err := sh.LoginUC.CheckAttempts(r.Context(), ip)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"ip":    ip,
			"error": "too many login attempts",
		}).Warn("login attempts limit exceeded")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("you have been temporary blocked, please try again at %s %v", blockTime, err)},
		)
		loginAttempts.WithLabelValues("false").Inc()
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"login": input.Login,
	}).Debug("attempting to create session")

	session, err := sh.LoginUC.CreateSession(r.Context(), usecase.LogInInput{
		Login:    input.Login,
		Password: input.Password,
	})
	// if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
	if err == pgx.ErrNoRows {
		MakeEasyJSONResponse(w, http.StatusOK,
			&model.ErrorResponse{Message: fmt.Sprintf("%v", err)},
		)
		return
	}

	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"login": input.Login,
			"ip":    ip,
			"error": err.Error(),
		}).Warn("failed to create session")

		sh.LoginUC.IncreaseAttempts(r.Context(), ip)

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("%v", err)},
		)
		loginAttempts.WithLabelValues("false").Inc()
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":    session.UserId,
		"session_id": session.SessionId,
	}).Info("user authenticated successfully")

	cookie, err := CreateCookies(session)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"user_id": session.UserId,
			"error":   err.Error(),
		}).Error("failed to create session cookie")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Failed to create cookie"},
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		HttpOnly: cookie.HttpOnly,
		Secure:   cookie.Secure,
		Expires:  cookie.Expires,
		Path:     cookie.Path,
		SameSite: http.SameSiteLaxMode,
	})

	token, _ := sh.LoginUC.CreateJwtToken(&repository.Session{
		ID:     session.SessionId,
		UserID: uint32(session.UserId),
	}, time.Now().Add(12*time.Hour).Unix())

	sh.Logger.WithFields(&logrus.Fields{
		"user_id": session.UserId,
	}).Debug("JWT token created")

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		HttpOnly: false,
		Secure:   false,
		Path:     "/",
		Expires:  time.Now().Add(12 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	_ = sh.LoginUC.DeleteAttempts(r.Context(), ip)

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":    session.UserId,
		"session_id": session.SessionId,
		"ip":         ip,
	}).Info("login completed successfully")

	loginAttempts.WithLabelValues("true").Inc()
	MakeEasyJSONResponse(w, http.StatusOK, &model.LoginResponse{
		Message: "Logged in",
		UserID:  session.UserId,
	})
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.SignUpRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Failed to read request body"},
		)
		return
	}

	if err := req.UnmarshalJSON(body); err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		return
	}

	user := req.User
	profile := req.Profile

	if uh.SignupUC.ValidateLogin(user.Login) != nil || uh.SignupUC.ValidatePassword(user.Password) != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid login or password"},
		)
		return
	}

	if uh.SignupUC.UserExists(user.Login) {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "User already exists"},
		)
		return
	}

	profileId, err := uh.SignupUC.SaveUserProfile(profile)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Failed to save user profile %v", err)},
		)
		return
	}

	if _, err := uh.SignupUC.SaveUserData(profileId, user); err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Failed to save user data"},
		)
		return
	}

	MakeEasyJSONResponse(w, http.StatusCreated,
		&model.ErrorResponse{Message: "User created"},
	)
}

func (sh *SessionHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
	sh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("CheckSession request started")

	sessionChecks.WithLabelValues("inactive").Inc()

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		sh.Logger.WithFields(&logrus.Fields{
			"error": "no session cookie found",
		}).Debug("session check failed - no cookies")

		response := model.SessionCheckResponse{
			Message:   "No cookies got",
			InSession: false,
		}
		MakeEasyJSONResponse(w, http.StatusOK, response)

		return
	} else if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Warn("failed to get session cookie")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid cookie"},
		)
		return
	}

	userId, err := sh.CheckSessionUC.CheckSession(session.Value)
	if err != nil {
		status := http.StatusInternalServerError
		message := "unknown session error"

		switch err {
		case model.ErrSessionNotFound:
			message = "session not found"
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Warn(message)

		case model.ErrGetSession:
			message = "error getting session"
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Error(message)

		case model.ErrInvalidSessionId:
			message = "error invalid session id"
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Warn(message)

		default:
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Error(message)
		}

		MakeEasyJSONResponse(w, status,
			&model.ErrorResponse{Message: message},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":    userId,
		"session_id": session.Value,
	}).Info("session check successful")

	response := model.SessionCheckSuccessResponse{
		Message:   "Logged in",
		InSession: true,
		UserId:    userId,
	}

	sessionChecks.WithLabelValues("active").Inc()
	MakeEasyJSONResponse(w, http.StatusOK, response)
}

func (sh *SessionHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	sh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("Logout request started")

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		sh.Logger.WithFields(&logrus.Fields{
			"error": "session cookie not found",
		}).Warn("logout attempt without session cookie")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "No cookies got"},
		)
		return
	} else if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Error("failed to get session cookie")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid cookie"},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"session_id": session.Value,
	}).Debug("attempting to logout session")

	if err := sh.LogoutUC.Logout(session.Value); err != nil {
		if err == model.ErrSessionNotFound {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Warn("session not found during logout")

			MakeEasyJSONResponse(w, http.StatusInternalServerError,
				&model.ErrorResponse{Message: "session not found"},
			)

			logoutAttempts.WithLabelValues("false").Inc()
			return
		}
		if err == model.ErrGetSession {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Error("failed to get session during logout")

			MakeEasyJSONResponse(w, http.StatusInternalServerError,
				&model.ErrorResponse{Message: "error getting session"},
			)
			logoutAttempts.WithLabelValues("false").Inc()
			return
		}
		if err == model.ErrDeleteSession {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Error("failed to delete session")

			MakeEasyJSONResponse(w, http.StatusInternalServerError,
				&model.ErrorResponse{Message: "error deleting session"},
			)
			logoutAttempts.WithLabelValues("false").Inc()
			return
		}

		sh.Logger.WithFields(&logrus.Fields{
			"session_id": session.Value,
			"error":      err.Error(),
		}).Error("unknown logout error")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "unknown logout error"},
		)
		logoutAttempts.WithLabelValues("false").Inc()
		return
	}

	expiredCookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().AddDate(-1, 0, 0),
		Path:     "/",
	}
	http.SetCookie(w, expiredCookie)

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().AddDate(-1, 0, 0),
		Path:     "/",
	})

	sh.Logger.WithFields(&logrus.Fields{
		"session_id": session.Value,
	}).Info("user logged out successfully")

	logoutAttempts.WithLabelValues("true").Inc()

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: "Logged out"},
	)
}

func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	uh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("DeleteUser request started")

	sanitizer := bluemonday.UGCPolicy()
	vars := mux.Vars(r)
	id := vars["id"]

	uh.Logger.WithFields(&logrus.Fields{
		"raw_user_id": id,
	}).Debug("received user ID for deletion")

	userId, err := strconv.Atoi(sanitizer.Sanitize(id))
	if err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"raw_user_id": id,
			"error":       err.Error(),
		}).Warn("invalid user ID format")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid user id"},
		)
		return
	}

	uh.Logger.WithFields(&logrus.Fields{
		"user_id": userId,
	}).Info("attempting to delete user")

	if err := uh.DeleteUserUC.DeleteUser(userId); err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"user_id": userId,
			"error":   err.Error(),
		}).Error("failed to delete user")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Error deleting user"},
		)
		return
	}

	uh.Logger.WithFields(&logrus.Fields{
		"user_id": userId,
	}).Info("user deleted successfully")

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: fmt.Sprintf("User with ID %d deleted", userId)},
	)
}

func (uh *UserHandler) GetUserParams(w http.ResponseWriter, r *http.Request) {
	uh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetUserParams request started")
	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint32)
	if !ok {
		uh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Info("unauthorized profile access attempt")
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	uh.Logger.Info("attempting to get user params")

	user, err := uh.GetParamsUC.GetUserParams(int(userID))
	uh.Logger.Info("Error getting user: ", err)

	if err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting user: %v", err)},
		)
		return
	}

	uh.Logger.Info("user params received successfully")
	MakeEasyJSONResponse(w, http.StatusOK, user)
}

func (ph *ProfilesHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetProfile request started")

	profileRetrieved.WithLabelValues("get profile").Inc()
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized profile access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Debug("attempting to get profile")

	profile, err := ph.GetProfileUC.GetProfile(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get profile")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting profile: %v", err)},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Info("profile retrieved successfully")

	is_admin, err := ph.GetAdminUC.GetAdmin(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"user_id": profileId,
			"error":   err.Error(),
		}).Error("failed to get answers for query")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions for query: %v", err)},
		)
		return
	}

	profileIsAdmin := model.ProfileIsAdmin{
		ProfileId:   profile.ProfileId,
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		IsMale:      profile.IsMale,
		Goal:        profile.Goal,
		Height:      profile.Height,
		Birthday:    profile.Birthday,
		Description: profile.Description,
		Location:    profile.Location,
		Interests:   profile.Interests,
		LikedBy:     profile.LikedBy,
		Preferences: profile.Preferences,
		Parameters:  profile.Parameters,
		Photos:      profile.Photos,
		Premium:     profile.Premium,
		IsAdmin:     is_admin,
	}

	MakeEasyJSONResponse(w, http.StatusOK, profileIsAdmin)
}

func (ph *ProfilesHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetProfiles request started")

	profilesListRetrieved.WithLabelValues("get profiles").Inc()
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)

	IsPremiumRaw := r.Context().Value(isPremiumKey)
	IsPremium, _ := IsPremiumRaw.(bool)

	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized profiles access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"requester_id": profileId,
	}).Debug("attempting to get profiles list")

	redisKey := fmt.Sprintf("cached_profiles:%d", profileId)

	cached, err := ph.Subscriber.Exists(context.Background(), redisKey).Result()
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Redis error"},
		)
		return
	}
	if cached > 0 {
		cachedData, err := ph.Subscriber.Get(context.Background(), redisKey).Result()
		if err == nil {
			var cachedProfiles []model.Profile
			if err := json.Unmarshal([]byte(cachedData), &cachedProfiles); err == nil {
				MakeEasyJSONResponse(w, http.StatusOK, model.ProfileResponse{Profiles: cachedProfiles})
				return
			}
		}
	}

	if !IsPremium {
		viewKey := fmt.Sprintf("profile_view_limit:%d", profileId)
		countStr, err := ph.Subscriber.Get(context.Background(), viewKey).Result()
		if err != nil && err != redis.Nil {
			MakeEasyJSONResponse(w, http.StatusInternalServerError,
				&model.ErrorResponse{Message: "Redis error"},
			)
			return
		}

		viewCount := 0
		if countStr != "" {
			viewCount, _ = strconv.Atoi(countStr)
		}

		if viewCount >= model.MaxProfileViewsWithoutSub {
			MakeEasyJSONResponse(w, http.StatusAccepted,
				&model.ErrorResponse{Message: "Go touch grass"},
			)
			return
		}

		pipe := ph.Subscriber.TxPipeline()
		pipe.Incr(context.Background(), viewKey)
		if viewCount == 0 {
			pipe.Expire(context.Background(), viewKey, 12*time.Hour)
		}
		_, _ = pipe.Exec(context.Background())
	}

	profiles, err := ph.GetProfilesUC.GetProfiles(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"requester_id": profileId,
			"error":        err.Error(),
		}).Error("failed to get profiles list")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting profiles: %v", err)},
		)
		return
	}

	data, err := json.Marshal(profiles)
	if err == nil {
		_ = ph.Subscriber.Set(context.Background(), redisKey, data, 30*time.Minute).Err()
	}

	ph.Logger.WithFields(&logrus.Fields{
		"requester_id":   profileId,
		"profiles_count": len(profiles),
	}).Info("profiles list retrieved successfully")

	MakeEasyJSONResponse(w, http.StatusOK, model.ProfileResponse{Profiles: profiles})
}

func (ph *ProfilesHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":       r.Method,
		"path":         r.URL.Path,
		"request_id":   r.Header.Get("request_id"),
		"ip":           r.RemoteAddr,
		"query_params": r.URL.Query(),
	}).Info("DeletePhoto request started")

	photoRemoved.WithLabelValues("delete photo").Inc()
	sanitizer := bluemonday.UGCPolicy()
	fileURL := sanitizer.Sanitize(r.URL.Query().Get("file_url"))

	ph.Logger.WithFields(&logrus.Fields{
		"raw_file_url":  r.URL.Query().Get("file_url"),
		"sanitized_url": fileURL,
	}).Debug("processing file URL")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized photo deletion attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"user_id":  user_id,
		"file_url": fileURL,
	}).Info("attempting to delete photo")

	err := ph.DeleteImageUC.DeleteImage(int(user_id), fileURL)
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"user_id":  user_id,
			"file_url": fileURL,
			"error":    err.Error(),
		}).Error("failed to delete photo")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error deleting photo: %v", err)},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"user_id":  user_id,
		"file_url": fileURL,
	}).Info("photo deleted successfully")

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: fmt.Sprintf("Deleted photo %s for user %d", fileURL, user_id)},
	)
}

func (qh *QueryHandler) GetActiveQueries(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetActiveQueries request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to get active queries")

	queries, err := qh.GetActiveQueriesUC.GetActiveQueries(int32(user_id))
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get active queries")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting active queries: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("active queries retrieved successfully")

	MakeEasyJSONResponse(w, http.StatusOK, model.QuerResponse{Queries: queries})
}

func (qh *QueryHandler) StoreUserAnswer(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("SendUserAnswer request started")

	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	var answer model.UserAnswer
	if err := easyjson.UnmarshalFromReader(r.Body, &answer); err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("failed to decode answer")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("Error decoding answer: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": userID,
		"answer":  answer,
	}).Info("attempting to store user answer")

	err := qh.StoreUserAnswerUC.StoreUserAnswer(int32(userID), answer.Name, answer.Score, answer.Answer)
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"answer":  answer,
			"error":   err.Error(),
		}).Error("failed to store user answer")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error storing user answer: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": userID,
		"answer":  answer,
	}).Info("user answer stored successfully")

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: "User answer stored successfully"},
	)
}

func (qh *QueryHandler) GetAnswersForUser(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetAnswersForUser request started")

	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": userID,
	}).Info("attempting to get answers for user")

	answers, err := qh.GetAnswersForUserUC.GetAnswersForUser(int32(userID))
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("failed to get answers for user")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting answers for user: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": userID,
	}).Info("answers for user retrieved successfully")

	MakeEasyJSONResponse(w, http.StatusOK, model.QueryResponse{Queries: answers})
}

func (qh *QueryHandler) FindQuery(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetAnswersForQuery request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	var req model.FindQueryRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to decode FindQueryRequest")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("Error decoding request: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to get answers for query")

	is_admin, err := qh.GetAdminUC.GetAdmin(int(user_id))
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get answers for query")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions for query: %v", err)},
		)
		return
	}

	if !is_admin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You dont have permissions"},
		)
		return
	}

	answers, err := qh.FindQueryUC.FindQuery(req.Name, req.Query_id)
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get answers for query")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting answers for query: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("answers for query retrieved successfully")

	MakeEasyJSONResponse(w, http.StatusOK, model.AnswersForResponse{Answers: answers})
}

func (qh *QueryHandler) DeleteQuery(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("DeleteQuery request started")

	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.Warn("unauthorized query access attempt: missing or invalid userID in context")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	var req model.DeleteQueryRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Error("failed to decode FindQueryRequest")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("Error decoding request: %v", err)},
		)
		return
	}

	isAdmin, err := qh.GetAdminUC.GetAdmin(int(userID))
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("failed to get admin permissions")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions: %v", err)},
		)
		return
	}

	if !isAdmin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have permissions"},
		)
		return
	}

	err = qh.DeleteQueryUC.DeleteAnswer(req.User_id, req.Query_name)
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("failed to delete query answer")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error deleting query answer: %v", err)},
		)
		return
	}

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: "Deleted successfully"},
	)
}

func (qh *QueryHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetStatistics request started")

	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.Warn("unauthorized query access attempt: missing or invalid userID in context")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	var answer model.GetAnswerStatistics
	if err := easyjson.UnmarshalFromReader(r.Body, &answer); err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Error("failed to decode FindQueryRequest")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("Error decoding request: %v", err)},
		)
		return
	}

	isAdmin, err := qh.GetAdminUC.GetAdmin(int(userID))
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("failed to get admin permissions")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions: %v", err)},
		)
		return
	}

	if !isAdmin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have permissions"},
		)
		return
	}

	stats, err := qh.GetStatisticsUC.GetStatistics(answer.Query_name)
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("failed to get statistics")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting statistics: %v", err)},
		)
		return
	}

	MakeEasyJSONResponse(w, http.StatusOK, stats)
}

func (qh *QueryHandler) GetAnswersForQuery(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetAnswersForQuery request started")

	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.Warn("unauthorized query access attempt: missing or invalid userID in context")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	isAdmin, err := qh.GetAdminUC.GetAdmin(int(userID))
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("failed to get admin permissions")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions: %v", err)},
		)
		return
	}

	if !isAdmin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have permissions"},
		)
		return
	}

	answers, err := qh.GetAnswersForQueryUC.GetAnswersForQuery()
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("failed to get answers for query")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting answers for query: %v", err)},
		)
		return
	}

	MakeEasyJSONResponse(w, http.StatusOK, model.AnswersResponse{Answers: answers})
}

func (ch *ComplaintHandler) CreateComplaint(w http.ResponseWriter, r *http.Request) {
	ch.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("CreateComplaint request started")

	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint32)
	if !ok {
		ch.Logger.Warn("unauthorized query access attempt: missing or invalid userID in context")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		ch.Logger.WithError(err).Warn("failed to read request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid request body"},
		)
		return
	}
	defer r.Body.Close()

	var req model.CreateComplaintRequest
	if err := req.UnmarshalJSON(body); err != nil {
		ch.Logger.WithError(err).Warn("invalid JSON in request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		return
	}

	var complOn int
	if req.Complaint_on == "" {
		complOn = 0
	} else {
		complOn, err = strconv.Atoi(req.Complaint_on)
		if err != nil {
			ch.Logger.WithError(err).Warn("invalid complaint_on value")

			MakeEasyJSONResponse(w, http.StatusBadRequest,
				&model.ErrorResponse{Message: "Invalid complaint_on value"},
			)
			return
		}
	}

	if err := ch.CreateComplateUC.CreateComplaint(int(userID), complOn, req.Complaint_type, req.Complaint_text); err != nil {
		ch.Logger.WithError(err).Error("failed to create complaint")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Failed to save complaint"},
		)
		return
	}

	MakeEasyJSONResponse(w, http.StatusCreated,
		&model.ErrorResponse{Message: "Complaint created"},
	)
}

func (ch *ComplaintHandler) GetComplaints(w http.ResponseWriter, r *http.Request) {
	ch.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetComplaints request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		ch.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	ch.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to get complaints")

	is_admin, err := ch.GetAdminUC.GetAdmin(int(user_id))
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get complaints")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions for query: %v", err)},
		)
		return
	}

	if !is_admin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: fmt.Sprintf("You dont have permissions: %v", err)},
		)
		return
	}

	complaints, err := ch.GetComplaintsUC.GetAllComplaints()
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get complaints")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting complaint: %v", err)},
		)
		return
	}

	ch.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("complaints retrieved successfully")

	MakeEasyJSONResponse(w, http.StatusOK, model.ComplaintsResponse{Complaints: complaints})
}

func (sh *SubHandler) AddSubscription(w http.ResponseWriter, r *http.Request) {
	sh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("CreateComplaint request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		sh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	var label string
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			MakeEasyJSONResponse(w, http.StatusBadRequest,
				&model.ErrorResponse{Message: "Failed to read request body"},
			)
			return
		}

		var req model.AddSubRequet
		lexer := jlexer.Lexer{Data: body}
		req.UnmarshalEasyJSON(&lexer)
		if lexer.Error() != nil {
			MakeEasyJSONResponse(w, http.StatusBadRequest,
				&model.ErrorResponse{Message: "Invalid JSON"},
			)
			return
		}
		label = req.Label
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			MakeEasyJSONResponse(w, http.StatusBadRequest,
				&model.ErrorResponse{Message: "Invalid form"},
			)
			return
		}

		label = r.FormValue("label")
	} else {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid content type"},
		)
		return
	}

	sub_id, err := strconv.Atoi(label)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid label format"},
		)
		return
	}

	var builder strings.Builder
	for key, values := range r.Form {
		if key == "label" {
			continue
		}
		for _, v := range values {
			builder.WriteString(fmt.Sprintf("%s=%s; ", key, v))
		}
	}
	combined := builder.String()

	if err := sh.AddSubUC.CreateSub(int(user_id), sub_id, combined); err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Failed to save user data"},
		)
		return
	}

	MakeEasyJSONResponse(w, http.StatusCreated,
		&model.ErrorResponse{Message: "Subsr created"},
	)
}

func (sh *SubHandler) ChangeBorder(w http.ResponseWriter, r *http.Request) {
	sh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("SetLike request started")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		sh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	IsPremiumRaw := r.Context().Value(isPremiumKey)
	IsPremium, _ := IsPremiumRaw.(bool)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Warn("failed to read request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Failed to read request body"},
		)
		return
	}

	var input model.ChangeBorderRequest
	lexer := jlexer.Lexer{Data: body}
	input.UnmarshalEasyJSON(&lexer)
	if lexer.Error() != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      lexer.Error().Error(),
		}).Warn("failed to decode like request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		return
	}

	if !IsPremium {
		sh.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"new_border": input.NewBorder,
		}).Warn("no premium")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "You cannot update border with no subscription"},
		)
		return
	}

	err = sh.UpdateBorderUC.UpdateBorder(int(profileId), input.NewBorder)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"new_border": input.NewBorder,
			"error":      err.Error(),
		}).Error("failed to set like")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error updating border: %v", err)},
		)
		return
	}

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: "Changed"},
	)
}

func (ch *ComplaintHandler) FindComplaint(w http.ResponseWriter, r *http.Request) {
	ch.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetComplaints request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		ch.Logger.Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	is_admin, err := ch.GetAdminUC.GetAdmin(int(user_id))
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: "Error getting admin permissions"},
		)
		return
	}

	if !is_admin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have permissions"},
		)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		ch.Logger.Warn("failed to read request body")
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Failed to read request body"},
		)
		return
	}

	if len(body) == 0 {
		complaints, err := ch.GetComplaintsUC.GetAllComplaints()
		if err != nil {
			MakeEasyJSONResponse(w, http.StatusInternalServerError,
				&model.ErrorResponse{Message: "Error getting complaints"},
			)
			return
		}
		MakeEasyJSONResponse(w, http.StatusOK, model.ComplaintsResponse{Complaints: complaints})
		return
	}

	var input model.FindComplaint

	lexer := jlexer.Lexer{Data: body}
	input.UnmarshalEasyJSON(&lexer)
	if lexer.Error() != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid complaint filter"},
		)
		return
	}

	complaints, err := ch.FindCompaintUC.FindComplaint(
		input.Complaint_by,
		input.Name_by,
		input.Complaint_on,
		input.Name_on,
		input.Complaint_type,
		input.Status,
	)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting complaints: %v", err)},
		)
		return
	}

	MakeEasyJSONResponse(w, http.StatusOK, model.ComplaintsResponse{Complaints: complaints})
}

func (ch *ComplaintHandler) DeleteComplaint(w http.ResponseWriter, r *http.Request) {
	ch.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("DeleteComplaint request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		ch.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to read request body")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "failed to read request body"},
		)
		return
	}

	var input model.DeleteComlaint

	lexer := jlexer.Lexer{Data: body}
	input.UnmarshalEasyJSON(&lexer)
	if lexer.Error() != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   lexer.Error(),
		}).Error("failed to decode complaint delete input")

		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("Error decoding complaint delete input: %v", lexer.Error())},
		)
		return
	}

	ch.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to delete complaint")

	is_admin, err := ch.GetAdminUC.GetAdmin(int(user_id))
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get admin permissions")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions: %v", err)},
		)
		return
	}

	if !is_admin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have permissions"},
		)
		return
	}

	err = ch.DeleteComplaintsUC.DeleteComplaint(input.Complaint_id)
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to delete complaint")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error deleting complaint: %v", err)},
		)
		return
	}

	ch.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("complaint deleted successfully")

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: "Deleted successful"},
	)
}

func (ch *ComplaintHandler) HandleComplaint(w http.ResponseWriter, r *http.Request) {
	ch.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("HandleComplaint request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		ch.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "failed to read request body"},
		)
		return
	}

	var input model.HandleComplaint
	lexer := jlexer.Lexer{Data: body}
	input.UnmarshalEasyJSON(&lexer)
	if lexer.Error() != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   lexer.Error(),
		}).Error("failed to decode complaint input")
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: fmt.Sprintf("Error decoding complaint input: %v", lexer.Error())},
		)
		return
	}

	ch.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to handle complaint")

	is_admin, err := ch.GetAdminUC.GetAdmin(int(user_id))
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get admin permissions")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions: %v", err)},
		)
		return
	}

	if !is_admin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have permissions"},
		)
		return
	}

	err = ch.HandleComplaintUC.HandleComplaint(input.Complaint_id, input.NewStatus)
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to update complaint")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error updating complaint: %v", err)},
		)
		return
	}

	ch.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("complaint updated successfully")

	MakeEasyJSONResponse(w, http.StatusOK,
		&model.ErrorResponse{Message: "Updated successful"},
	)
}

func (ch *ComplaintHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	ch.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetAnswersForQuery request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		ch.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	var raw model.RawTimeConstraints
	body, err := io.ReadAll(r.Body)
	if err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "failed to read request body"},
		)
		return
	}
	if err := raw.UnmarshalJSON(body); err != nil {
		MakeEasyJSONResponse(w, http.StatusBadRequest,
			&model.ErrorResponse{Message: "Invalid JSON"},
		)
		return
	}

	var constraints model.TimeConstraints
	var useTimeFrom, useTimeTo bool

	if raw.TimeFrom != nil {
		t, err := time.Parse(time.RFC3339, *raw.TimeFrom)
		if err != nil {
			http.Error(w, "invalid time_from format", http.StatusBadRequest)
			return
		}
		constraints.TimeFrom = t
		useTimeFrom = true
	}

	if raw.TimeTo != nil {
		t, err := time.Parse(time.RFC3339, *raw.TimeTo)
		if err != nil {
			http.Error(w, "invalid time_to format", http.StatusBadRequest)
			return
		}
		constraints.TimeTo = t
		useTimeTo = true
	}

	ch.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to get statistics for complaints")

	is_admin, err := ch.GetAdminUC.GetAdmin(int(user_id))
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get answers for query")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting admin permissions for complaints: %v", err)},
		)
		return
	}

	if !is_admin {
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: fmt.Sprintf("You dont have permissions: %v", err)},
		)
		return
	}

	stats, err := ch.GetStatisticsUC.GetStatistics(useTimeFrom, constraints.TimeFrom, useTimeTo, constraints.TimeTo)
	if err != nil {
		ch.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get answers for query")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting statistics for complaints : %v", err)},
		)

		return
	}

	ch.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("statistics for complaints retrieved successfully")

	MakeEasyJSONResponse(w, http.StatusOK, stats)
}

func (ph *ProfilesHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetProfile request started")

	profileRetrieved.WithLabelValues("get profile").Inc()
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.Warn("unauthorized profile access attempt")
		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	cacheKey := fmt.Sprintf("recommendation:%d", profileId)
	lockKey := fmt.Sprintf("recommendation_lock:%d", profileId)

	exists, err := ph.Subscriber.Exists(context.Background(), lockKey).Result()
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("Redis error on lock check")
	}

	if exists > 0 {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
		}).Warn("recommendation requested too often")

		MakeEasyJSONResponse(w, http.StatusTooManyRequests,
			&model.ErrorResponse{Message: "Recommendations can be requested only once per 24 hours"},
		)
		return
	}

	cached, err := ph.Subscriber.Get(context.Background(), cacheKey).Bytes()
	if err == nil && len(cached) > 0 {
		var cachedProfile model.Profile
		if err := json.Unmarshal(cached, &cachedProfile); err == nil {
			MakeEasyJSONResponse(w, http.StatusOK, cachedProfile)
			return
		}
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Debug("fetching recommendation from DB")

	profile, err := ph.GetRecommendationsUC.GetRecommendations(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get recommendation")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting profile: %v", err)},
		)
		return
	}

	profileBytes, _ := json.Marshal(profile)
	if err := ph.Subscriber.Set(context.Background(), cacheKey, profileBytes, 24*time.Hour).Err(); err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Warn("failed to cache recommendation")
	}

	if err := ph.Subscriber.Set(context.Background(), lockKey, "1", 24*time.Hour).Err(); err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Warn("failed to set recommendation lock")
	}

	MakeEasyJSONResponse(w, http.StatusOK, profile)
}

func (ph *ProfilesHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetProfile request started")

	profileRetrieved.WithLabelValues("get profile").Inc()
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized profile access attempt")

		MakeEasyJSONResponse(w, http.StatusUnauthorized,
			&model.ErrorResponse{Message: "You don't have access"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Debug("attempting to get profile")

	stats, err := ph.GetProfileStatsUC.GetProfileStats(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get profile")

		MakeEasyJSONResponse(w, http.StatusInternalServerError,
			&model.ErrorResponse{Message: fmt.Sprintf("Error getting profile: %v", err)},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Info("profile retrieved successfully")

	MakeEasyJSONResponse(w, http.StatusOK, stats)
}
